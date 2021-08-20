package main

import (
	"gitee.com/kirile/kapi/k/cmd"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/os/gcmd"
	"github.com/linxlib/logs"
	"os"
)

var kVersion string
var goVersion string

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

func main() {
	defer func() {
		if exception := recover(); exception != nil {
			if err, ok := exception.(error); ok {
				_log.Print(gerror.Current(err).Error())
			} else {
				panic(exception)
			}
		}
	}()
	command := gcmd.GetArg(1)
	switch command {
	case "init", "i":
		cmd.Initialize()
	case "build":
		cmd.Build()
	case "install":
		cmd.Install()
	case "run":
		cmd.Run()
	default:
		cmd.Install()
	}
}
