package cmd

import (
	"fmt"
	"github.com/linxlib/kapi/cmd/k/utils"
	"github.com/linxlib/kapi/cmd/k/utils/innerlog"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	buildFileContent = `# 编译配置
[k]
   name = "%s" # 编译的可执行文件名
   version = "1.0.0" # 版本
   arch = "amd64"  # 平台
   system = "windows,linux" # 系统
   path = "./bin" # 输出目录
`

	mainContent = `package main
import (
	"github.com/linxlib/kapi"
)
func main() {
	k := kapi.New()
	//k.RegisterRouter(new(controller.CategoryController))
	k.Run()
}
`
	dockerFileContent = `
FROM golang:1.18 as build
MAINTAINER "author <email>"
ARG MODE=prod
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    TZ=Asia/Shanghai \
    GOOS=linux \
    GOARCH=amd64 \
	GOPROXY="https://goproxy.cn" \
	GOPRIVATE="gitee.com"

RUN mkdir /src
WORKDIR /src
ADD go.mod .
ADD go.sum .
RUN go mod download

COPY . .
RUN make all MODE=${MODE}

FROM saranraj/alpine-tz-ca-certificates as prod
ARG MODE=prod
ENV TZ=Asia/Shanghai
RUN mkdir /app
WORKDIR /app

COPY --from=build /src/bin/<appname> .
COPY --from=build /src/cmd/<appname>/config/config_$MODE.toml ./config.toml
RUN ln -fs /app/<appname> /usr/bin/<appname>
EXPOSE 1404

CMD ["<appname>"]
`

	defaultConf = `
[server]
debug = true
needDoc = true
docName = "K-Api"
docDesc = "K-Api"
port = 2022
docDomain = ""
docVer = "v1"
redirectToDocWhenAccessRoot = true
enablePProf = true
[[staticDirs]]
    dir = "static"
    path = "static"
[server.cors]
allowHeaders = ["Origin","Content-Length","Content-Type"]
`
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "create kapi project layout in current directory",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var modName string
		if len(args) > 0 {
			modName = args[0]
			exec.Command("go", "mod", "init", modName)
		} else {
			if !utils.Exists("go.mod") {
				innerlog.Log.Println("go.mod不存在")
				return
			}
			modName = utils.GetMod("go.mod")
		}
		i := strings.LastIndex(modName, "/")
		mod := ""
		if i > 0 {
			mod = modName[i:]
		} else {
			mod = modName
		}

		if !utils.Exists("cmd") {
			utils.Mkdir("cmd/" + mod)
		}

		if !utils.Exists("config") {
			utils.Mkdir("config")
		}

		if !utils.Exists("config/build.toml") {
			r := fmt.Sprintf(buildFileContent, modName)
			ioutil.WriteFile("config/build.toml", []byte(r), os.ModePerm)
			innerlog.Log.Println("写出build.toml")
		}
		if !utils.Exists("config/config.toml") {
			ioutil.WriteFile("config/config.toml", []byte(defaultConf), os.ModePerm)
			innerlog.Log.Println("写出config.toml")
		}
		if !utils.Exists("api") {
			innerlog.Log.Println("创建api目录")
			utils.Mkdir("api")
			utils.Mkdir("api/controller")
			utils.Mkdir("api/service")
			utils.Mkdir("api/model")
		}
		if !utils.Exists("pkg") {
			innerlog.Log.Println("创建pkg目录")
			utils.Mkdir("pkg")
		}
		if !utils.Exists("main.go") {
			//r := fmt.Sprintf(mainContent, modName, modName)
			ioutil.WriteFile("main.go", []byte(mainContent), os.ModePerm)
			innerlog.Log.Println("写出main.go")
			fmtcmd := exec.Command("gofmt", "-l", "-w", "./")
			fmtcmd.Stdout = os.Stderr

			if err := fmtcmd.Start(); err != nil {
				innerlog.Log.Errorln(err)
			}
			if err := fmtcmd.Wait(); err != nil {
				innerlog.Log.Errorln(err)
			}
			c := exec.Command("go", "mod", "tidy")
			c.Stdout = os.Stderr

			if err := c.Start(); err != nil {
				innerlog.Log.Errorln(err)
			}
			if err := c.Wait(); err != nil {
				innerlog.Log.Errorln(err)
			}

		}
		if !utils.Exists("Dockerfile") {
			a := strings.ReplaceAll(dockerFileContent, "<appname>", modName)
			ioutil.WriteFile("Dockerfile", []byte(a), os.ModePerm)
			innerlog.Log.Println("写出Dockerfile")
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//initCmd.Flags().StringP("toggle", "t", false, "Help message for toggle")
}
