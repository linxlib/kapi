package cmd

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
)

func GenProto() {
	parser, err := gcmd.Parse(g.MapStrBool{
		"f,file": false,
	})
	if err != nil {
		_log.Fatal(err)
		return
	}
	file := parser.GetArg(2)
	if len(file) < 1 {
		// Check and use the main.go file.
		if gfile.Exists("proto/*.proto") {

		} else {

		}
	}

}
