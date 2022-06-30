package kapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/linxlib/conf"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/cors"
	"time"
)

// RecoverErrorFunc recover 错误设置
type RecoverErrorFunc func(interface{})

type Option struct {
	ginLoggerFormatter gin.LogFormatter
	corsConfig         cors.Config
	recoverErrorFunc   RecoverErrorFunc
	intranetIP         string

	Server struct {
		Debug                       bool     `conf:"debug"`
		NeedDoc                     bool     `conf:"needDoc"`
		NeedReDoc                   bool     `conf:"needReDoc"`
		NeedSwagger                 bool     `conf:"needSwagger"`
		DocName                     string   `conf:"docName" default:"K-Api"`
		DocDesc                     string   `conf:"docDesc" default:"K-Api"`
		Port                        int      `conf:"port" default:"2022"`
		DocVer                      string   `conf:"docVer" default:"v1"`
		RedirectToDocWhenAccessRoot bool     `conf:"redirectToDocWhenAccessRoot"`
		StaticDirs                  []string `conf:"staticDirs" default:"[static]"`
		EnablePProf                 bool     `conf:"enablePProf"`
		Cors                        struct {
			AllowAllOrigins     bool          `conf:"allowAllOrigins"`
			AllowCredentials    bool          `conf:"allowCredentials"`
			MaxAge              time.Duration `conf:"maxAge" default:"12h"`
			AllowWebSockets     bool          `conf:"allowWebSockets"`
			AllowWildcard       bool          `conf:"allowWildcard"`
			AllowPrivateNetwork bool          `conf:"allowPrivateNetwork"`
			AllowMethods        []string      `conf:"allowMethods" default:"[GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS]"`
			AllowHeaders        []string      `conf:"allowHeaders" default:"[Origin,Content-Length,Content-Type,Authorization,x-requested-with]"`
		} `conf:"cors"`
	} `conf:"server"`
}

func readConfig(o *Option) *Option {
	//配置cors
	corsConfig := cors.DefaultConfig()
	o.Server.Cors.AllowAllOrigins = true
	o.Server.Debug = true
	o.Server.NeedDoc = true
	o.Server.NeedSwagger = true
	o.Server.NeedReDoc = false
	o.Server.Cors.AllowPrivateNetwork = true
	o.Server.Cors.AllowWebSockets = true
	o.Server.EnablePProf = false
	_ = conf.Load(o, conf.File("config.toml"),
		conf.Dirs("./", "./config"))

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
				internal.ByteCountSI(int64(param.BodySize)),
				param.ErrorMessage,
			)
		},
		intranetIP: internal.GetIntranetIp(),
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
