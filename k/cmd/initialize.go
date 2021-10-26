package cmd

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gfile"
	"golang.org/x/mod/modfile"
	"io/ioutil"
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
		if gfile.Exists("config.toml") {
			if !g.Config().Contains("k.name") {
				g.Config().Set("k.name",modName)
			}
			if !g.Config().Contains("k.version") {
				g.Config().Set("k.version","1.0.0")
			}
			if !g.Config().Contains("k.arch") {
				g.Config().Set("k.arch","amd64")
			}
			if !g.Config().Contains("k.system") {
				g.Config().Set("k.system","windows,linux")
			}
			if !g.Config().Contains("k.path") {
				g.Config().Set("k.path","./bin")
			}
		}
	}


}
