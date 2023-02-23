package kapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/linxlib/config"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/cors"
	"net"
	"os"
	"time"
)

// RecoverFunc recover 错误设置
type RecoverFunc func(interface{})

type StaticDir struct {
	Path string `yaml:"path"`
	Root string `yaml:"root"`
}

type ServerOption struct {
	NeedDoc    bool        `yaml:"needDoc"`
	DocName    string      `yaml:"docName"`
	DocDesc    string      `yaml:"docDesc"`
	BasePath   string      `yaml:"basePath"`
	Port       int         `yaml:"port"`
	DocVer     string      `yaml:"docVer"`
	StaticDirs []StaticDir `yaml:"staticDirs"`
	Cors       cors.Config `yaml:"cors"`
}

var _defaultServerOption = ServerOption{
	NeedDoc:  true,
	DocName:  "KApi",
	DocDesc:  "KApi",
	BasePath: "",
	Port:     time.Now().Year(),
	DocVer:   "v1",
	StaticDirs: []StaticDir{
		{Path: "static", Root: "static"},
	},
	Cors: cors.DefaultConfig(),
}

type Option struct {
	ginLoggerFormatter gin.HandlerFunc
	recoverErrorFunc   RecoverFunc
	intranetIP         string
	corsHandler        gin.HandlerFunc
	y                  *config.YAML
	Server             ServerOption
}

func readConfig(o *Option) *Option {
	if internal.FileIsExist("config/config.yaml") {
		conf, err := config.NewYAML(config.File("config/config.yaml"))
		if err != nil {
			internal.Errorf("%s", err)
		}
		err = conf.Get("server").Populate(&_defaultServerOption)
		if err != nil {
			internal.Errorf("%s", err)
		}
		o.y = conf
	} else {
		internal.Warnf("file %s not exist, use default options", "config/config.yaml")
	}
	o.Server = _defaultServerOption
	o.corsHandler = cors.New(o.Server.Cors)

	return o
}

func defaultOption() *Option {
	o := &Option{
		ginLoggerFormatter: gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
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
				byteCountSI(int64(param.BodySize)),
				param.ErrorMessage,
			)
		}),
		intranetIP: getIntranetIP(),
		recoverErrorFunc: func(err interface{}) {
			switch err {
			case KAPIEXIT:
				return
			default:
				internal.Errorf("%s", err)
			}
		},
	}
	return readConfig(o)
}

// ByteCountSI 字节数转带单位
func byteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func getIntranetIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		internal.Errorf("%s", err)
		os.Exit(1)
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}

		}
	}
	return "localhost"
}

func (o *Option) SetGinLoggerFormatter(formatter gin.LogFormatter) *Option {
	o.ginLoggerFormatter = gin.LoggerWithFormatter(formatter)
	return o
}
func (o *Option) Get() *config.YAML {
	return o.y
}
func (o *Option) SetRecoverFunc(f func(interface{})) *Option {
	o.recoverErrorFunc = func(err interface{}) {
		switch err {
		case KAPIEXIT:
			return
		default:
			internal.Errorf("%s", err)
			f(err)
		}
	}
	return o
}

func (o *Option) SetIntranetIP(ip string) *Option {
	o.intranetIP = ip
	return o
}
