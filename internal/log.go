package internal

import (
	"github.com/linxlib/logs"
	"os"
)

var Log logs.FieldLogger = &logs.Logger{
	Out:   os.Stdout,
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

var ErrorLog logs.FieldLogger = &logs.Logger{
	Out:   os.Stdout,
	Hooks: make(logs.LevelHooks),
	Formatter: &logs.TextFormatter{
		ForceColors:      true,
		FullTimestamp:    true,
		TimestampFormat:  "01-02 15:04:05",
		QuoteEmptyFields: false,
		CallerPrettifier: nil,
		HideLevelText:    true,
	},
	ReportCaller: true,
	Level:        logs.DebugLevel,
}
