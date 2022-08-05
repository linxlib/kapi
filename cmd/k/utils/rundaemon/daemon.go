package rundaemon

import (
	"bufio"
	"fmt"
	"github.com/linxlib/kapi/cmd/k/utils/innerlog"
	"github.com/linxlib/logs"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Milliseconds to wait for the next job to begin after a file change
const WorkDelay = 900

// Default pattern to match files which trigger a build
const FilePattern = `(.+\.go|.+\.c)$`

type globList []string

func (g *globList) String() string {
	return fmt.Sprint(*g)
}
func (g *globList) Set(value string) error {
	*g = append(*g, filepath.Clean(value))
	return nil
}
func (g *globList) Matches(value string) bool {
	for _, v := range *g {
		if match, err := filepath.Match(v, value); err != nil {
			log.Fatalf("Bad pattern \"%s\": %s", v, err.Error())
		} else if match {
			return true
		}
	}
	return false
}

type buildCommandList struct {
	commands []string
}

func (f *buildCommandList) String() string {
	return ""
}

func (f *buildCommandList) Set(s string) error {
	f.commands = append(f.commands, s)
	return nil
}

var (
	flagBuildCommandList buildCommandList
	flagDirectories      globList
	flagExcludedDirs     globList
	flagExcludedFiles    globList
	flagIncludedFiles    globList
	flagBuildDir         string
	flagRunDir           string
	flagCommandStop      bool
	flagCommand          string
	flagGracefulKill     bool
	flagGracefulTimeout  time.Duration
	flagPattern          string
)

func Run(args ...string) {
	flagCommand = "./" + buildOutputName
	flagPattern = FilePattern
	flagGracefulKill = false
	a := []string{"go", "build", "-o", buildOutputName}
	a = append(a, args...)
	s := strings.Join(a, " ")
	flagBuildCommandList.Set(s)
	flagCommandStop = false
	if len(flagDirectories) == 0 {
		flagDirectories = globList([]string{"."})
	}
	if flagGracefulKill && !gracefulTerminationPossible() {
		innerlog.Log.Fatal("Graceful termination is not supported on your platform.")
	}
	pattern := regexp.MustCompile(flagPattern)
	cfg := &WatcherConfig{
		flagVerbose:       false,
		flagRecursive:     true,
		flagDirectories:   flagDirectories,
		flagExcludedDirs:  flagExcludedDirs,
		flagExcludedFiles: flagExcludedFiles,
		flagIncludedFiles: flagIncludedFiles,
		pattern:           pattern,
	}
	watcher, err := NewWatcher(cfg)
	if err != nil {
		innerlog.Log.Fatal(err)
	}

	defer watcher.Close()

	err = watcher.AddFiles()
	if err != nil {
		innerlog.Log.Fatal("watcher.Addfiles():", err)
	}

	jobs := make(chan string)
	buildSuccess := make(chan bool)
	buildStarted := make(chan string)

	go builder(jobs, buildStarted, buildSuccess)

	if flagCommand != "" {
		go runner(flagCommand, buildStarted, buildSuccess)
	} else {
		go flusher(buildStarted, buildSuccess)
	}

	watcher.Watch(jobs) // start watching files
}

// Run `go build` and print the output if something's gone wrong.
func build() bool {
	innerlog.Log.Infoln("building")

	for _, c := range flagBuildCommandList.commands {
		err := runBuildCommand(c)
		if err != nil {
			innerlog.Log.Errorln("command failed: ", c)
			return false
		}
	}
	innerlog.Log.Infoln("build ok.")

	return true
}
func runBuildCommand(c string) error {
	c = strings.TrimSpace(c)
	args := strings.Split(c, " ")
	if len(args) == 0 {
		return nil
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = flagDirectories[0]
	if flagBuildDir != "" {
		cmd.Dir = flagBuildDir
	} else if len(flagDirectories) > 0 {

	}

	output, err := cmd.CombinedOutput()

	if err != nil {
		innerlog.Log.Errorln("build failed:\n", string(output))
		return err
	}
	return nil
}
func matchesPattern(pattern *regexp.Regexp, file string) bool {
	return pattern.MatchString(file)
}

// Accept build jobs and start building when there are no jobs rushing in.
// The inrush protection is WorkDelay milliseconds long, in this period
// every incoming job will reset the timer.
func builder(jobs <-chan string, buildStarted chan<- string, buildDone chan<- bool) {
	createThreshold := func() <-chan time.Time {
		return time.After(time.Duration(WorkDelay * time.Millisecond))
	}

	threshold := createThreshold()
	eventPath := ""

	for {
		select {
		case eventPath = <-jobs:
			threshold = createThreshold()
		case <-threshold:
			buildStarted <- eventPath
			buildDone <- build()
		}
	}
}
func logger(pipeChan <-chan io.ReadCloser) {
	dumper := func(pipe io.ReadCloser, prefix string) {
		reader := bufio.NewReader(pipe)

	readloop:
		for {
			line, err := reader.ReadString('\n')

			if err != nil {
				break readloop
			}
			innerlog.Log.Print(line)
		}
	}

	for {
		pipe := <-pipeChan
		go dumper(pipe, "stdout:")

		pipe = <-pipeChan
		go dumper(pipe, "stderr:")
	}
}

// Start the supplied command and return stdout and stderr pipes for logging.
func startCommand(command string) (cmd *exec.Cmd, stdout io.ReadCloser, stderr io.ReadCloser, err error) {
	args := strings.Split(command, " ")
	cmd = exec.Command(args[0], args[1:]...)

	if flagRunDir != "" {
		cmd.Dir = flagRunDir
	}

	if stdout, err = cmd.StdoutPipe(); err != nil {
		err = fmt.Errorf("can't get stdout pipe for command: %s", err)
		return
	}

	if stderr, err = cmd.StderrPipe(); err != nil {
		err = fmt.Errorf("can't get stderr pipe for command: %s", err)
		return
	}

	if err = cmd.Start(); err != nil {
		err = fmt.Errorf("can't start command: %s", err)
		return
	}

	return
}

// Run the command in the given string and restart it after
// a message was received on the buildDone channel.
func runner(commandTemplate string, buildStarted <-chan string, buildSuccess <-chan bool) {
	var currentProcess *os.Process
	pipeChan := make(chan io.ReadCloser)

	go logger(pipeChan)

	// Launch concurrent process watching for signals from outside that
	// indicate termination to kill the running process alongside
	// CompileDaemon to prevent orphan processes.
	go func() {
		processSignalChannel := make(chan os.Signal, 1)
		signal.Notify(processSignalChannel, fatalSignals...)
		<-processSignalChannel

		logs.Infoln("received signal, exiting, over.")
		if currentProcess != nil {
			killProcess(currentProcess)
		}
		os.Exit(0)
	}()

	for {
		eventPath := <-buildStarted

		// prepend %0.s (which adds nothing) to prevent warning about missing
		// format specifier if the user did not supply one.
		command := fmt.Sprintf("%0.s"+commandTemplate, eventPath)

		if !flagCommandStop {
			if !<-buildSuccess {
				continue
			}
		}

		if currentProcess != nil {
			killProcess(currentProcess)
		}

		if flagCommandStop {
			innerlog.Log.Infoln("Command stopped. Waiting for build to complete.")
			if !<-buildSuccess {
				continue
			}
		}

		innerlog.Log.Infoln("starting.")
		innerlog.Log.Println("-------------------------------------------")
		cmd, stdoutPipe, stderrPipe, err := startCommand(command)

		if err != nil {
			innerlog.Log.Errorf("Could not start command: %s", err)
		}

		pipeChan <- stdoutPipe
		pipeChan <- stderrPipe

		currentProcess = cmd.Process
	}
}

func killProcess(process *os.Process) {
	if flagGracefulKill {
		killProcessGracefully(process)
	} else {
		killProcessHard(process)
	}
}

func killProcessHard(process *os.Process) {
	innerlog.Log.Infoln("Hard stopping the current process..")

	if err := process.Kill(); err != nil {
		innerlog.Log.Warnln("Warning: could not kill child process.  It may have already exited.")
	}

	if _, err := process.Wait(); err != nil {
		innerlog.Log.Fatalln("Could not wait for child process. Aborting due to danger of infinite forks.", err)
	}
}

func killProcessGracefully(process *os.Process) {
	done := make(chan error, 1)
	go func() {
		innerlog.Log.Infoln("Gracefully stopping the current process..")
		if err := terminateGracefully(process); err != nil {
			done <- err
			return
		}
		_, err := process.Wait()
		done <- err
	}()

	select {
	case <-time.After(flagGracefulTimeout * time.Second):
		innerlog.Log.Infoln("Could not gracefully stop the current process, proceeding to hard stop.")
		killProcessHard(process)
		<-done
	case err := <-done:
		if err != nil {
			innerlog.Log.Fatalln("Could not kill child process. Aborting due to danger of infinite forks.")
		}
	}
}

func flusher(buildStarted <-chan string, buildSuccess <-chan bool) {
	for {
		<-buildStarted
		<-buildSuccess
	}
}
