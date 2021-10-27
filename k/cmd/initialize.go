package cmd

import (
	"fmt"
	"github.com/gogf/gf/os/gfile"
	"golang.org/x/mod/modfile"
	"io/ioutil"
	"os"
)

var (
	buildFileContent = `# 编译配置
[k]
   name = "%s" # 编译的可执行文件名
   version = "1.0.0" # 版本
   arch = "amd64"  # 平台
   system = "windows,linux" # 系统
   path = "./bin" # 输出目录
`
	configFileContent = `# mysql 数据库配置
[db]
   enable = false 
   mysql = "root:root@tcp(127.0.0.1:3306)/test?charset=utf8"

# redis 配置
[redis]
   enable = false
   address = "127.0.0.1:6379"
   password = ""
   db = 1
`
)

func Initialize() {
	//TODO: 写出默认的配置文件
	if gfile.Exists("go.mod") {
		modName := ""
		bs, _ := ioutil.ReadFile("go.mod")
		f, _ := modfile.Parse("go.mod", bs, func(_, version string) (string, error) {
			return version, nil
		})
		modName =  f.Module.Mod.String()
		if !gfile.Exists("build.toml") {
			r := fmt.Sprintf(buildFileContent,modName)
			ioutil.WriteFile("build.toml",[]byte(r),os.ModePerm)
			_log.Println("写出build.toml")
		}
		if !gfile.Exists("config.toml") {
			//r := fmt.Sprintf(configFileContent,modName)
			ioutil.WriteFile("config.toml",[]byte(configFileContent),os.ModePerm)
			_log.Println("写出config.toml")
		}
		if !gfile.Exists("api") {
			gfile.Mkdir("api")
		}
	} else {
		_log.Println("go.mod不存在")
	}
}
