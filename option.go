package kapi

import (
	"gitee.com/kirile/kapi/internal"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
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
	outPath                     string
	ginLoggerFormatter          gin.LogFormatter
	corsConfig                  cors.Config
	apiBasePath                 string

	listenPort int

	recoverErrorFunc RecoverErrorFunc
	intranetIP       string

	staticDir string
}

func defaultOption() *Option {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AddAllowHeaders("Access-Control-Allow-Private-Network: true")
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	return &Option{
		isDebug:                     true,
		needDoc:                     true,
		docName:                     "K-Api",
		openDocInBrowser:            false,
		docDomain:                   "",
		docVer:                      "v1",
		redirectToDocWhenAccessRoot: true,
		docDesc:                     "K-Api",
		apiBasePath:                 "/",
		listenPort:                  2021,
		outPath:                     "",
		ginLoggerFormatter:          defaultLogFormatter,
		corsConfig:                  corsConfig,
		intranetIP:                  internal.GetIntranetIp(),
		staticDir: "",
		recoverErrorFunc: func(err interface{}) {
			switch err {
			case KAPIEXIT:
				return
			default:
				_log.Error(err)
			}
		},
	}
}

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
		o.isDebug = needDoc[0]
	}
	return o
}

func (o *Option) SetDocName(docName string) *Option {
	o.docName = docName
	return o
}
func (o *Option) SetDocVersion(ver string) *Option {
	o.docVer = ver
	return o
}
func (o *Option) SetDocDescription(desc string) *Option {
	o.docDesc = desc
	return o
}
func (o *Option) SetOutputPath(outPath string) *Option {
	if !strings.HasSuffix(outPath, "/") {
		outPath += "/"
	}
	o.outPath = outPath
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

func (o *Option) SetGinLoggerFormater(formatter gin.LogFormatter) *Option {
	o.ginLoggerFormatter = formatter
	return o
}

func (o *Option) SetRecoverFunc(f func(interface{})) *Option {
	o.recoverErrorFunc = func(err interface{}) {
		switch err {
		case KAPIEXIT:
			return
		default:
			_log.Error(err)
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

func (o *Option) SetStaticDir(dir ...string) *Option {
	o.staticDir = "static"
	if len(dir) > 0 {
		o.staticDir = dir[0]
	}
	return o

}
