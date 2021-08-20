package cmd

import (
	"fmt"
	"github.com/gogf/gf/container/garray"
	"github.com/gogf/gf/container/gtype"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gfsnotify"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/gf/text/gstr"
	"github.com/linxlib/logs"
	"os"
	"runtime"
)

var _log logs.FieldLogger = &logs.Logger{
	Out:   os.Stderr,
	Hooks: make(logs.LevelHooks),
	Formatter: &logs.TextFormatter{
		ForceColors:      true,
		FullTimestamp:    true,
		TimestampFormat:  "01-02 15:04:05",
		QuoteEmptyFields: false,
		CallerPrettyfier: nil,
		HideLevelText:    true,
	},
	ReportCaller: false,
	Level:        logs.TraceLevel,
	//ReportFunc:   false,
}
var (
	process *gproc.Process
	kill    = gtype.NewBool(false)
)

type App struct {
	File    string // Go run file name/path.
	Options string // Extra "go run" options.
	Args    string // run args
}

func (app *App) Run() {
	renamePath := ""
	//_log.Printf("build: %s", app.File)
	outputPath := gfile.Join("bin", gfile.Name(app.File))
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
		if gfile.Exists(outputPath) {
			renamePath = outputPath + "~"
			if err := gfile.Rename(outputPath, renamePath); err != nil {
				_log.Error(err)
			}
		}
	}
	buildCommand := fmt.Sprintf(`go build -o %s %s %s`, outputPath, app.Options, app.File)
	//_log.Print(buildCommand)
	result, err := gproc.ShellExec(buildCommand)
	if err != nil {
		_log.Errorf("编译失败: \n%s%s", result, err.Error())
		return
	}
	if process != nil {
		if err := process.Kill(); err != nil {
			_log.Errorf("[k run]停止进程失败: %s", err.Error())
			//return
		}
	}
	// Run the binary file.
	runCommand := fmt.Sprintf(`%s %s`, outputPath, app.Args)
	//_log.Print(runCommand)
	if runtime.GOOS == "windows" {
		// Special handling for windows platform.
		// DO NOT USE "cmd /c" command.
		process = gproc.NewProcess(runCommand, nil)
	} else {
		process = gproc.NewProcessCmd(runCommand, nil)
	}
	if pid, err := process.Start(); err != nil {
		_log.Errorf("启动失败: %s", err.Error())
	} else {
		_log.Infof("进程ID: %d", pid)
	}
	if kill.Val() {
		kill.Set(false)
	}
	if err := process.Wait(); err != nil {
		if kill.Val() {
			return
		}
		_log.Fatalf("[k run] 监听进程异常退出: %v", err)
	}
}

func Run() {

	gproc.AddSigHandlerShutdown(func(sig os.Signal) {
		_log.Printf("[k run] 退出 %v", sig)
		if process != nil {
			process.Kill()
		}
		os.Exit(0)
	})
	go gproc.Listen()
	parser, err := gcmd.Parse(g.MapStrBool{
		"args": true,
	})
	if err != nil {
		_log.Fatal(err)
	}

	file := gcmd.GetArg(2)
	if len(file) < 1 {
		file = "main.go"
	}
	app := &App{
		File: file,
	}
	array := garray.NewStrArrayFrom(os.Args)
	args := parser.GetOpt("args")
	if args != "" {
		app.Args = args
		index := -1
		array.Iterator(func(k int, v string) bool {
			if gstr.Contains(v, "-args") {
				index = k
				return false
			}
			return true
		})
		if index != -1 {
			v, _ := array.Get(index)
			if gstr.Contains(v, "=") {
				array.Remove(index)
			} else {
				array.Remove(index)
				array.Remove(index)
			}
		}
	}
	dirty := gtype.NewBool()
	_, err = gfsnotify.Add(gfile.RealPath("."), func(event *gfsnotify.Event) {
		if gfile.ExtName(event.Path) != "go" {
			return
		}
		if gstr.Contains(gfile.Name(event.Path), "gen_router") {
			return
		}
		// Variable `dirty` is used for running the changes only one in one second.
		if !dirty.Cas(false, true) {
			return
		}
		// With some delay in case of multiple code changes in very short interval.
		gtimer.SetTimeout(1500*gtime.MS, func() {
			_log.Printf(`go源码变更: %s`, event.String())
			kill.Set(true)
			dirty.Set(false)
			app.Run()
		})
	})
	if err != nil {
		_log.Fatal(err)
	}

	go app.Run()
	select {}

}
