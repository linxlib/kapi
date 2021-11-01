package cmd

import (
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"golang.org/x/mod/modfile"
	"io/ioutil"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const platforms = `
    darwin    amd64
    darwin    arm64
    ios       amd64
    ios       arm64
    freebsd   386
    freebsd   amd64
    freebsd   arm
    linux     386
    linux     amd64
    linux     arm
    linux     arm64
    linux     ppc64
    linux     ppc64le
    linux     mips
    linux     mipsle
    linux     mips64
    linux     mips64le
    netbsd    386
    netbsd    amd64
    netbsd    arm
    openbsd   386
    openbsd   amd64
    openbsd   arm
    windows   386
    windows   amd64
	android   arm
	dragonfly amd64
	plan9     386
	plan9     amd64
	solaris   amd64
`

func Build() {
	g.Config().SetFileName("build.toml")
	parser, err := gcmd.Parse(g.MapStrBool{
		"n,name":    true,
		"v,version": true,
		"a,arch":    true,
		"s,system":  true,
		"p,path":    true,
	})
	if err != nil {
		_log.Fatal(err)
	}
	file := parser.GetArg(2)
	if len(file) < 1 {
		// Check and use the main.go file.
		if gfile.Exists("main.go") {
			file = "main.go"
		} else {
			_log.Fatal("编译文件不能为空")
		}
	}
	path := getOption(parser, "path", "./bin")
	name := getOption(parser, "name", gfile.Name(file))
	if len(name) < 1 || name == "*" {
		_log.Fatal("名称不能为空")
	}

	var (
		version       = getOption(parser, "version")
		archOption    = getOption(parser, "arch")
		systemOption  = getOption(parser, "system")
		customSystems = gstr.SplitAndTrim(systemOption, ",")
		customArches  = gstr.SplitAndTrim(archOption, ",")
	)

	if len(version) > 0 {
		path += "/" + version
	}
	// System and arch checks.
	var (
		spaceRegex  = regexp.MustCompile(`\s+`)
		platformMap = make(map[string]map[string]bool)
	)
	for _, line := range strings.Split(strings.TrimSpace(platforms), "\n") {
		line = gstr.Trim(line)
		line = spaceRegex.ReplaceAllString(line, " ")
		var (
			array  = strings.Split(line, " ")
			system = strings.TrimSpace(array[0])
			arch   = strings.TrimSpace(array[1])
		)
		if platformMap[system] == nil {
			platformMap[system] = make(map[string]bool)
		}
		platformMap[system][arch] = true
	}
	modName := ""
	if gfile.Exists("./go.mod") {
		bs, _ := ioutil.ReadFile("go.mod")
		f, _ := modfile.Parse("go.mod", bs, func(_, version string) (string, error) {
			return version, nil
		})
		modName = f.Module.Mod.String()

	}

	ldFlags := ""

	// start building
	_log.Print("开始编译...")
	genv.Set("CGO_ENABLED", "0")
	var (
		cmd = ""
		ext = ""
	)
	for system, item := range platformMap {
		cmd = ""
		ext = ""
		if len(customSystems) > 0 && customSystems[0] != "all" && !gstr.InArray(customSystems, system) {
			continue
		}
		for arch, _ := range item {
			if len(customArches) > 0 && customArches[0] != "all" && !gstr.InArray(customArches, arch) {
				continue
			}
			if len(customSystems) == 0 && len(customArches) == 0 {
				if runtime.GOOS == "windows" {
					ext = ".exe"
				}
				ldFlags = fmt.Sprintf(`-X gitee.com/kirile/kapi.VERSION=%s`, "NO_VERSION") +
					fmt.Sprintf(` -X gitee.com/kirile/kapi.BUILDTIME=%s`, time.Now().Format("2006-01-02T15:04:01")) +
					fmt.Sprintf(` -X gitee.com/kirile/kapi.GOVERSION=%s`, runtime.Version()) +
					fmt.Sprintf(` -X gitee.com/kirile/kapi.OS=%s`, runtime.GOOS) +
					fmt.Sprintf(` -X gitee.com/kirile/kapi.ARCH=%s`, runtime.GOARCH) +
					fmt.Sprintf(` -X gitee.com/kirile/kapi.PACKAGENAME=%s`, modName)
				// Single binary building, output the binary to current working folder.
				output := "-o " + name + ext
				cmd = fmt.Sprintf(`go build %s -ldflags "%s"  %s`, output, ldFlags, file)
			} else {
				// Cross-building, output the compiled binary to specified path.
				if system == "windows" {
					ext = ".exe"
				}
				genv.Set("GOOS", system)
				genv.Set("GOARCH", arch)
				ldFlags = fmt.Sprintf(`-X gitee.com/kirile/kapi.VERSION=%s`, version) +
					fmt.Sprintf(` -X gitee.com/kirile/kapi.BUILDTIME=%s`, time.Now().Format("2006-01-02T15:04:01")) +
					fmt.Sprintf(` -X gitee.com/kirile/kapi.GOVERSION=%s`, runtime.Version()) +
					fmt.Sprintf(` -X gitee.com/kirile/kapi.OS=%s`, system) +
					fmt.Sprintf(` -X gitee.com/kirile/kapi.ARCH=%s`, arch) +
					fmt.Sprintf(` -X gitee.com/kirile/kapi.PACKAGENAME=%s`, modName)
				cmd = fmt.Sprintf(
					`go build -o %s/%s/%s%s -ldflags "%s" %s`,
					path, system+"_"+arch, name, ext, ldFlags, file,
				)
			}
			_log.Debug(cmd)
			// It's not necessary printing the complete command string.
			cmdShow, _ := gregex.ReplaceString(`\s+(-ldflags ".+?")\s+`, " ", cmd)
			_log.Print(cmdShow)
			if result, err := gproc.ShellExec(cmd); err != nil {
				_log.Printf("编译失败, os:%s, arch:%s, error:\n%s\n", system, arch, gstr.Trim(result))
			}
			if gfile.Exists("gen.gob") {
				gfile.CopyFile("gen.gob", fmt.Sprintf(
					`%s/%s/gen.gob`,
					path, system+"_"+arch))
				_log.Debug("拷贝gen.gob文件")
			}
			if gfile.Exists("swagger.json") {
				gfile.CopyFile("swagger.json", fmt.Sprintf(
					`%s/%s/swagger.json`,
					path, system+"_"+arch))
				_log.Debug("拷贝swagger.json文件")
			}
			if gfile.Exists("config.toml") && !gfile.Exists(fmt.Sprintf(
				`%s/%s/config.toml`,
				path, system+"_"+arch)) {
				gfile.CopyFile("config.toml", fmt.Sprintf(
					`%s/%s/config.toml`,
					path, system+"_"+arch))
				_log.Debug("拷贝config.toml文件")
			}
			// single binary building.
			if len(customSystems) == 0 && len(customArches) == 0 {

				goto buildDone
			}
		}
	}
buildDone:
	_log.Print("完成!")

}

const nodeNameInConfigFile = "k"

// getOption retrieves option value from parser and configuration file.
// It returns the default value specified by parameter `value` is no value found.
func getOption(parser *gcmd.Parser, name string, value ...string) (result string) {
	result = parser.GetOpt(name)
	if result == "" && g.Config().Available() {
		result = g.Config().GetString(nodeNameInConfigFile + "." + name)
	}
	if result == "" && len(value) > 0 {
		result = value[0]
	}
	return
}
