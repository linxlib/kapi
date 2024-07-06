/*
Copyright © 2023 linx sulinke1133@gmail.com
*/
package cmd

import (
	"fmt"
	"github.com/linxlib/kapi/cmd/k/utils"
	"github.com/linxlib/logs"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build",
	Example: `
k build windows amd64 example                //make windows executable
k build windows,darwin amd64 example         //make windows and Mac OS X executables
k build windows amd64 example -v 1.2.0       //make windows executable with kapi.VERSION and other variables be set, which will be printed out when running`,
	Long: `build the project with configs or flags. make executable files for many platforms with only one line command.`,
	Run: func(cmd *cobra.Command, args []string) {
		modName := utils.GetMod("go.mod")
		// TODO: validate parameters
		//args[0] system
		//args[1] arch
		//args[2] executable name. default is mod name. eg. if mod name is github.com/linxlib/example, the executable name will be `example` `example.exe` etc..
		count := len(args)
		switch count {
		case 1:
			systemStr := args[0]
			archStr := "amd64"
			i := strings.LastIndex(modName, "/")
			var name string
			if i < 0 {
				name = modName
			} else {
				name = modName[i:]
			}
			out := cmd.Flag("out").Value.String()
			version := cmd.Flag("version").Value.String()
			tags := cmd.Flag("tags").Value.String()
			//ldflags := cmd.Flag("ldflags").Value.String()

			obj := buildObject{
				system: strings.Split(systemStr, ","),
				arch:   strings.Split(archStr, ","),
				//ldflags: ldflags,
				name:    name,
				out:     out,
				tags:    tags,
				modName: modName,
				version: version,
			}
			err := doBuild(obj)
			if err != nil {
				logs.Error(err)
				return
			}

		case 2:
			systemStr := args[0]
			archStr := args[1]

			i := strings.LastIndex(modName, "/")
			var name string
			if i < 0 {
				name = modName
			} else {
				name = modName[i:]
			}

			out := cmd.Flag("out").Value.String()
			version := cmd.Flag("version").Value.String()
			tags := cmd.Flag("tags").Value.String()
			//ldflags := cmd.Flag("ldflags").Value.String()

			obj := buildObject{
				system: strings.Split(systemStr, ","),
				arch:   strings.Split(archStr, ","),
				name:   name,
				out:    out,
				//ldflags: ldflags,
				tags:    tags,
				modName: modName,
				version: version,
			}
			err := doBuild(obj)
			if err != nil {
				logs.Error(err)
				return
			}
		case 3:
			systemStr := args[0]
			archStr := args[1]
			name := args[2]
			if name == "" {
				i := strings.LastIndex(modName, "/")
				name = modName[i:]
			}

			out := cmd.Flag("out").Value.String()
			version := cmd.Flag("version").Value.String()
			tags := cmd.Flag("tags").Value.String()
			//ldflags := cmd.Flag("ldflags").Value.String()
			obj := buildObject{
				system: strings.Split(systemStr, ","),
				arch:   strings.Split(archStr, ","),
				name:   name,
				out:    out,
				//ldflags: ldflags,
				modName: modName,
				tags:    tags,
				version: version,
			}
			err := doBuild(obj)
			if err != nil {
				logs.Error(err)
				return
			}
		default:
			logs.Error("specify a target system OS parameter at least.")
		}

	},
}

type buildObject struct {
	system  []string
	arch    []string
	ldflags string
	modName string
	name    string
	out     string
	tags    string
	version string
}

func doBuild(obj buildObject) error {
	hostVersion, hostSystem, hostArch := utils.GetGoVersion()
	cmd := []string{"build", "-o"}
	ext := ""
	for _, s := range obj.system {
		if s == "windows" {
			ext = ".exe"
		}
	}
	for _, s := range obj.system {
		for _, arch := range obj.arch {
			logs.Infof("building %s_%s", s, arch)
			_ = os.Setenv("GOOS", s)
			_ = os.Setenv("GOARCH", arch)
			outPath := fmt.Sprintf("%s/%s_%s/%s%s%s", obj.out, s, arch, obj.version+"/", obj.name, ext)
			cmd = append(cmd, outPath)
			//cmd = append(cmd, "-ldflags="+obj.ldflags)
			if obj.tags != "" {
				cmd = append(cmd, "-tags="+obj.tags)
			}

			if obj.version != "" {
				cmd = append(cmd, "-ldflags")
				var tmp string
				tmp += fmt.Sprintf("-X 'github.com/linxlib/kapi.VERSION=%s'", obj.version)
				tmp += fmt.Sprintf(" -X 'github.com/linxlib/kapi.BUILDTIME=%s'", time.Now().Format(time.DateTime))
				tmp += fmt.Sprintf(" -X 'github.com/linxlib/kapi.GOVERSION=%s'", hostVersion)
				tmp += fmt.Sprintf(" -X 'github.com/linxlib/kapi.BUILDOS=%s'", hostSystem)
				tmp += fmt.Sprintf(" -X 'github.com/linxlib/kapi.BUILDARCH=%s'", hostArch)
				tmp += fmt.Sprintf(" -X 'github.com/linxlib/kapi.OS=%s'", s)
				tmp += fmt.Sprintf(" -X 'github.com/linxlib/kapi.ARCH=%s'", arch)
				tmp += fmt.Sprintf(" -X 'github.com/linxlib/kapi.PACKAGENAME=%s'", obj.modName)
				cmd = append(cmd, tmp)

			}
			err := utils.RunCommand("go", cmd...)
			if err != nil {
				logs.Error(err)
				return err
			}
		}

	}
	return nil
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	buildCmd.Flags().StringP("out", "o", "bin", "output director. default is ./bin")
	buildCmd.Flags().StringP("ldflags", "l", "", "golang ldflags.")
	buildCmd.Flags().StringP("tags", "t", "", "golang build tags.")
	buildCmd.Flags().StringP("version", "v", "1.0.0", "version. will set kapi.VERSION")
}
