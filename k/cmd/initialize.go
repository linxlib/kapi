package cmd

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gfile"
)

func Initialize() {
	//TODO: 写出默认的配置文件
	if !gfile.Exists("config.toml") {
		g.Config().Set("k.name","api_base")
		g.Config().Set("k.version","1.0.0")
		g.Config().Set("k.arch","amd64")
		g.Config().Set("k.system","windows,linux")
		g.Config().Set("k.path","./bin")
	}

}
