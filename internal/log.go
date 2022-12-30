package internal

import (
	"github.com/linxlib/logs"
	"os"
)

var Log logs.FieldLogger = &logs.Logger{
	Out:   os.Stderr,
	Hooks: make(logs.LevelHooks),
	Formatter: &logs.TextFormatter{
		ForceColors:      true,
		FullTimestamp:    true,
		TimestampFormat:  "01-02 15:04:05",
		QuoteEmptyFields: false,
		CallerPrettifier: nil,
		HideLevelText:    true,
	},
	ReportCaller: false,
	Level:        logs.DebugLevel,
}
