package kapi

import (
	"gitee.com/kirile/kapi/internal"
	"gitee.com/kirile/kapi/internal/cors"
	"gitee.com/kirile/kapi/lib/toml"
	"github.com/gin-gonic/gin"
)

type Option struct {
	isDebug                     bool
	needDoc                     bool
	docName                     string
	openDocInBrowser            bool
	redirectToDocWhenAccessRoot bool
	docDomain                   string
	docDesc                     string
	docVer                      string
	ginLoggerFormatter          gin.LogFormatter
	corsConfig                  cors.Config
	apiBasePath                 string
	listenPort                  int
	recoverErrorFunc            RecoverErrorFunc
	intranetIP                  string
	staticDir                   []string
}

const SECTION_SERVER_NAME = "server"

var _config = toml.ParseFile("config.toml")

func readConfig(o *Option) *Option {
	//配置cors
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowPrivateNetwork = true

	arr := _config.Array(SECTION_SERVER_NAME+".cors.allowHeaders", toml.ParseDefaultArray(`["Origin","Content-Length","Content-Type","Authorization","x-requested-with"]`))
	allowHeaders := make([]string, 0)
	for _, value := range arr {
		allowHeaders = append(allowHeaders, value.AsString())
	}
	corsConfig.AllowHeaders = allowHeaders
	o.corsConfig = corsConfig
	o.SetIsDebug(_config.Bool(SECTION_SERVER_NAME+".debug", true))
	o.SetNeedDoc(_config.Bool(SECTION_SERVER_NAME+".needDoc", true))
	o.SetDocName(_config.String(SECTION_SERVER_NAME+".docName", "K-Api"))
	o.SetOpenDocInBrowser(_config.Bool(SECTION_SERVER_NAME+".openDocInBrowser", false))
	o.SetDocDomain(_config.String(SECTION_SERVER_NAME+".docDomain", ""))
	o.SetDocVersion(_config.String(SECTION_SERVER_NAME+".docVer", "v1"))
	o.SetRedirectToDocWhenAccessRoot(_config.Bool(SECTION_SERVER_NAME+".redirectToDocWhenAccessRoot", true))
	o.SetDocDescription(_config.String(SECTION_SERVER_NAME+".docDesc", "K-Api"))
	o.SetApiBasePath(_config.String(SECTION_SERVER_NAME+".apiBasePath", "/"))
	o.SetPort(_config.Int(SECTION_SERVER_NAME+".port", 2021))
	o.SetStaticDirs(_config.Strings(SECTION_SERVER_NAME + ".staticDirs")...)

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

// SetIsDebug 设置是否调试模式 当不是开发情况时自动变为false
func (o *Option) SetIsDebug(isDebug ...bool) *Option {
	o.isDebug = true
	if len(isDebug) > 0 {
		o.isDebug = isDebug[0]
	}
	if !internal.CheckFileIsExist("main.go") {
		o.isDebug = false
	}
	return o
}

func (o *Option) SetNeedDoc(needDoc ...bool) *Option {
	o.needDoc = true
	if len(needDoc) > 0 {
		o.needDoc = needDoc[0]
	}
	return o
}

// SetDocName 设置文档名称
func (o *Option) SetDocName(docName string) *Option {
	o.docName = docName
	return o
}

//SetDocVersion 设置文档版本
func (o *Option) SetDocVersion(ver string) *Option {
	o.docVer = ver
	return o
}

//SetDocDescription 设置文档描述
func (o *Option) SetDocDescription(desc string) *Option {
	o.docDesc = desc
	return o
}

func (o *Option) SetOpenDocInBrowser(open ...bool) *Option {
	o.openDocInBrowser = true
	if len(open) > 0 {
		o.openDocInBrowser = open[0]
	}
	return o
}

func (o *Option) SetDocDomain(docDomain string) *Option {
	o.docDomain = docDomain
	return o
}

func (o *Option) SetApiBasePath(path string) *Option {
	o.apiBasePath = path
	return o
}

func (o *Option) SetPort(port int) *Option {
	o.listenPort = port
	return o
}
func (o *Option) SetCors(cors cors.Config) *Option {
	o.corsConfig = cors
	return o
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

func (o *Option) SetRedirectToDocWhenAccessRoot(redirect ...bool) *Option {
	o.redirectToDocWhenAccessRoot = true
	if len(redirect) > 0 {
		o.redirectToDocWhenAccessRoot = redirect[0]
	}
	return o
}

func (o *Option) SetStaticDirs(dir ...string) *Option {
	o.staticDir = []string{"static"}
	if len(dir) > 0 {
		o.staticDir = dir
	}
	return o
}
