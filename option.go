package kapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/linxlib/conf"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/cors"
	"net"
	"os"
	"time"
)

// RecoverErrorFunc recover 错误设置
type RecoverErrorFunc func(interface{})

type Server struct {
	Debug                       bool     `conf:"debug"`
	NeedDoc                     bool     `conf:"needDoc"`
	DocName                     string   `conf:"docName" default:"K-Api"`
	DocDesc                     string   `conf:"docDesc" default:"K-Api"`
	Port                        int      `conf:"port" default:"2022"`
	DocVer                      string   `conf:"docVer" default:"v1"`
	RedirectToDocWhenAccessRoot bool     `conf:"redirectToDocWhenAccessRoot"`
	StaticDirs                  []string `conf:"staticDirs" default:"[static=static]"`
	EnablePProf                 bool     `conf:"enablePProf"`
	Cors                        struct {
		AllowAllOrigins     bool          `conf:"allowAllOrigins"`
		AllowCredentials    bool          `conf:"allowCredentials"`
		MaxAge              time.Duration `conf:"maxAge"`
		AllowWebSockets     bool          `conf:"allowWebSockets"`
		AllowWildcard       bool          `conf:"allowWildcard"`
		AllowPrivateNetwork bool          `conf:"allowPrivateNetwork"`
		AllowMethods        []string      `conf:"allowMethods" d`
		AllowHeaders        []string      `conf:"allowHeaders"`
	} `conf:"cors"`
}

type Option struct {
	ginLoggerFormatter gin.LogFormatter
	corsConfig         cors.Config
	recoverErrorFunc   RecoverErrorFunc
	intranetIP         string
	Server             Server `conf:"server"`
}

func readConfig(o *Option) *Option {
	//配置cors
	corsConfig := cors.DefaultConfig()
	o.Server.Cors.AllowAllOrigins = true
	o.Server.Debug = true
	o.Server.NeedDoc = true
	o.Server.Cors.AllowPrivateNetwork = true
	o.Server.Cors.AllowWebSockets = true
	o.Server.Cors.MaxAge = time.Hour * 12
	o.Server.Cors.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	o.Server.Cors.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	o.Server.EnablePProf = false
	_ = conf.Load(o, conf.File("config.toml"),
		conf.Dirs("./", "./config/"))

	corsConfig.AllowAllOrigins = o.Server.Cors.AllowAllOrigins
	corsConfig.AllowPrivateNetwork = o.Server.Cors.AllowPrivateNetwork
	corsConfig.AllowCredentials = o.Server.Cors.AllowCredentials
	corsConfig.MaxAge = o.Server.Cors.MaxAge
	corsConfig.AllowWebSockets = o.Server.Cors.AllowWebSockets
	corsConfig.AllowWildcard = o.Server.Cors.AllowWildcard
	corsConfig.AllowMethods = o.Server.Cors.AllowMethods
	corsConfig.AllowHeaders = o.Server.Cors.AllowHeaders
	o.corsConfig = corsConfig
	return o
}

func defaultOption() *Option {
	o := &Option{
		ginLoggerFormatter: func(param gin.LogFormatterParams) string {
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
		},
		intranetIP: getIntranetIP(),
		recoverErrorFunc: func(err interface{}) {
			switch err {
			case KAPIEXIT:
				return
			default:
				internal.Log.Error(err)
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
		internal.Log.Errorln(err)
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
	o.ginLoggerFormatter = formatter
	return o
}

func (o *Option) SetRecoverFunc(f func(interface{})) *Option {
	o.recoverErrorFunc = func(err interface{}) {
		switch err {
		case KAPIEXIT:
			return
		default:
			internal.Log.Error(err)
			f(err)
		}
	}
	return o
}

func (o *Option) SetIntranetIP(ip string) *Option {
	o.intranetIP = ip
	return o
}
