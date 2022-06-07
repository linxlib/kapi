package kapi

import (
	"github.com/gin-gonic/gin"
	"github.com/linxlib/conf"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/cors"
	"time"
)

type Option struct {
	ginLoggerFormatter gin.LogFormatter
	corsConfig         cors.Config
	recoverErrorFunc   RecoverErrorFunc
	intranetIP         string

	Server struct {
		Debug                       bool     `conf:"debug" default:"true"`
		NeedDoc                     bool     `conf:"needDoc" default:"true"`
		NeedReDoc                   bool     `conf:"needReDoc" default:"false"`
		NeedSwagger                 bool     `conf:"needSwagger" default:"true"`
		DocName                     string   `conf:"docName" default:"K-Api"`
		DocDesc                     string   `conf:"docDesc" default:"K-Api"`
		Port                        int      `conf:"port" default:"2022"`
		OpenDocInBrowser            bool     `conf:"openDocInBrowser" default:"true"`
		DocDomain                   string   `conf:"docDomain"`
		DocVer                      string   `conf:"docVer" default:"v1"`
		RedirectToDocWhenAccessRoot bool     `conf:"redirectToDocWhenAccessRoot" default:"true"`
		APIBasePath                 string   `conf:"apiBasePath" default:""`
		StaticDirs                  []string `conf:"staticDirs" default:"[static]"`
		EnablePProf                 bool     `conf:"enablePProf" default:"false"`
		Cors                        struct {
			AllowAllOrigins     bool          `conf:"allowAllOrigins" default:"true"`
			AllowCredentials    bool          `conf:"allowCredentials" default:"false"`
			MaxAge              time.Duration `conf:"maxAge" default:"12h"`
			AllowWebSockets     bool          `conf:"allowWebSockets" default:"true"`
			AllowWildcard       bool          `conf:"allowWildcard" default:"true"`
			AllowPrivateNetwork bool          `conf:"allowPrivateNetwork" default:"true"`
			AllowMethods        []string      `conf:"allowMethods" default:"[GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS]"`
			AllowHeaders        []string      `conf:"allowHeaders" default:"[Origin,Content-Length,Content-Type,Authorization,x-requested-with]"`
		} `conf:"cors"`
	} `conf:"server"`
}

func readConfig(o *Option) *Option {
	//配置cors
	corsConfig := cors.DefaultConfig()

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

	gin.ForceConsoleColor()
	o := &Option{
		ginLoggerFormatter: defaultLogFormatter,
		intranetIP:         internal.GetIntranetIp(),
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
