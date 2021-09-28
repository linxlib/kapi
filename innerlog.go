package kapi

import (
	"fmt"
	"gitee.com/kirile/kapi/internal"
	"github.com/gin-gonic/gin"
	"github.com/linxlib/logs"
	"os"
	"time"
)

var _log logs.FieldLogger = &logs.Logger{
	Out:   os.Stderr,
	Hooks: make(logs.LevelHooks),
	Formatter: &logs.TextFormatter{
		ForceColors:      true,
		FullTimestamp:    true,
		TimestampFormat:  "01-02 15:04:05.999",
		QuoteEmptyFields: false,
		CallerPrettyfier: nil,
		HideLevelText:    true,
	},
	ReportCaller: false,
	Level:        logs.TraceLevel,
}

var defaultLogFormatter = func(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		// Truncate in a golang < 1.8 safe way
		param.Latency = param.Latency - param.Latency%time.Second
	}
	return fmt.Sprintf("%v |%s %3d %s| %13v | %15s |%s %-7s %s %#v (%s)\n%s",
		param.TimeStamp.Format("01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		internal.ByteCountSI(int64(param.BodySize)),
		param.ErrorMessage,
	)
}
