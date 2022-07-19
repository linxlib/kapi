/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/linxlib/kapi/cmd/k/utils"
	"github.com/linxlib/kapi/cmd/k/utils/innerlog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/exec"
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

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build kapi to binary",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		name := viper.GetString("k.name")
		systemOption := viper.GetString("k.system")
		archOption := viper.GetString("k.arch")
		version := viper.GetString("k.version")
		path := viper.GetString("k.path")
		tags := viper.GetString("k.tags")

		customSystems := utils.SplitAndTrim(systemOption, ",")
		customArches := utils.SplitAndTrim(archOption, ",")

		if len(version) > 0 {
			path += "/" + version
		}
		var (
			spaceRegex  = regexp.MustCompile(`\s+`)
			platformMap = make(map[string]map[string]bool)
		)
		for _, line := range strings.Split(strings.TrimSpace(platforms), "\n") {
			line = utils.Trim(line)
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
		if utils.Exists("./go.mod") {
			modName = utils.GetMod("go.mod")
		}
		ldFlags := ""
		innerlog.Log.Print("开始编译...")

		if utils.Exists("config.toml") {
			innerlog.Log.Print("生成swagger...")
			cmd := exec.Command("go", "run", "main.go", "-g")

			if result, err := cmd.CombinedOutput(); err != nil {
				innerlog.Log.Printf("生成失败, error:\n%s\n", string(result))
			}
		} else {
			innerlog.Log.Print("一般golang项目, 跳过swagger生成")
		}

		os.Setenv("CGO_ENABLED", "0")
		var (
			ext = ""
		)
		for system, item := range platformMap {
			ext = ""
			if len(customSystems) > 0 && customSystems[0] != "all" && !utils.InArray(customSystems, system) {
				continue
			}
			for arch, _ := range item {
				if len(customArches) > 0 && customArches[0] != "all" && !utils.InArray(customArches, arch) {
					continue
				}
				if len(tags) > 0 {
					tags = "-tags=" + tags
				}
				if len(customSystems) == 0 && len(customArches) == 0 {
					if runtime.GOOS == "windows" {
						ext = ".exe"
					}
					ldFlags = fmt.Sprintf(`"-X github.com/linxlib/kapi.VERSION=%s`, "NO_VERSION") +
						fmt.Sprintf(` -X github.com/linxlib/kapi.BUILDTIME=%s`, time.Now().Format("2006-01-02T15:04:01")) +
						fmt.Sprintf(` -X github.com/linxlib/kapi.GOVERSION=%s`, runtime.Version()) +
						fmt.Sprintf(` -X github.com/linxlib/kapi.OS=%s`, runtime.GOOS) +
						fmt.Sprintf(` -X github.com/linxlib/kapi.ARCH=%s`, runtime.GOARCH) +
						fmt.Sprintf(` -X github.com/linxlib/kapi.PACKAGENAME=%s"`, modName)

				} else {
					// Cross-building, output the compiled binary to specified path.
					if system == "windows" {
						ext = ".exe"
					}
					os.Setenv("GOOS", system)
					os.Setenv("GOARCH", arch)
					ldFlags = fmt.Sprintf(`"-X github.com/linxlib/kapi.VERSION=%s`, version) +
						fmt.Sprintf(` -X github.com/linxlib/kapi.BUILDTIME=%s`, time.Now().Format("2006-01-02T15:04:01")) +
						fmt.Sprintf(` -X github.com/linxlib/kapi.GOVERSION=%s`, runtime.Version()) +
						fmt.Sprintf(` -X github.com/linxlib/kapi.OS=%s`, system) +
						fmt.Sprintf(` -X github.com/linxlib/kapi.ARCH=%s`, arch) +
						fmt.Sprintf(` -X github.com/linxlib/kapi.PACKAGENAME=%s"`, modName)

				}
				cmds := []string{
					"go",
					"build",
					"-o",
					path + "/" + system + "_" + arch + "/" + name + ext,
					"-ldflags",
					ldFlags,
				}
				if tags != "" {
					cmds = append(cmds, tags)
				}
				cmds = append(cmds, "main.go")

				shell := exec.Command(cmds[0], cmds[1:]...)
				//innerlog.Log.Println(shell.String())
				if result, err := shell.CombinedOutput(); err != nil {
					innerlog.Log.Errorf("编译失败, os:%s, arch:%s, error:\n%s\n", system, arch, string(result))
					continue
				}
				if utils.Exists("gen.gob") {
					utils.CopyFile("gen.gob", fmt.Sprintf(
						`%s/%s/gen.gob`,
						path, system+"_"+arch))
					innerlog.Log.Debug("拷贝gen.gob文件")
				}
				if utils.Exists("config.toml") && !utils.Exists(fmt.Sprintf(
					`%s/%s/config.toml`,
					path, system+"_"+arch)) {
					utils.CopyFile("config.toml", fmt.Sprintf(
						`%s/%s/config.toml`,
						path, system+"_"+arch))
					innerlog.Log.Debug("拷贝config.toml文件")
				}
				// single binary building.
				if len(customSystems) == 0 && len(customArches) == 0 {
					goto buildDone
				}
			}
		}
	buildDone:
		innerlog.Log.Print("完成!")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
