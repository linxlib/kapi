package kapi

import (
	"embed"
	"fmt"
	"gitee.com/kirile/kapi/ast"
	"gitee.com/kirile/kapi/doc/swagger"
	"gitee.com/kirile/kapi/internal"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"os"
	"reflect"
	"time"
)

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

// KApi base struct
type KApi struct {
	apiFun            ApiFunc
	customContextType reflect.Type

	beforeAfter Interceptor //拦截器
	engine      *gin.Engine
	baseGroup   *gin.RouterGroup
	doc         *swagger.DocSwagger
	option      *Option
}

// WithCtx use custom context.设置自定义context
//func WithCtx(middleware ApiFunc) Option {
//	return optionFunc(func(o *KApi) {
//		o.Model(middleware)
//	})
//}

// WithImportFile 添加自定义import文件列表
//func WithImportFile(k, v string) Option {
//	return optionFunc(func(o *KApi) {
//		ast.AddImportFile(k, v)
//	})
//}

// WithBeforeAfter 设置对象调用前后执行中间件
//func WithBeforeAfter(beforeAfter Interceptor) Option {
//	return optionFunc(func(o *KApi) {
//		o.beforeAfter = beforeAfter
//	})
//}

func New(f func(*Option)) *KApi {
	if VERSION != "" {
		_log.Infof(_banner, fmt.Sprintf(_info, VERSION, GOVERSION, OS, ARCH, BUILDTIME))
	} else {
		_log.Infof(_banner, "")
	}
	if PACKAGENAME != "" {
		_log.Infof("初始化[%s]", PACKAGENAME)
	} else {
		_log.Info("初始化..")
	}

	b := new(KApi)
	b.option = defaultOption()
	f(b.option)
	b.Model(NewAPIFunc)

	b.engine = gin.New()
	b.engine.Use(gin.LoggerWithFormatter(b.option.ginLoggerFormatter))
	b.engine.Use(cors.New(b.option.corsConfig))
	gin.SetMode(gin.ReleaseMode) //we don't need gin's debug output
	b.doc = swagger.NewDoc()

	_log.Infof("localIP:%s docName:%s port:%d", b.option.intranetIP, b.option.docName, b.option.listenPort)

	return b
}

func (b *KApi) RegisterRouter(cList ...interface{}) {
	_log.Debug("注册路由..")
	if b.option.needDoc {
		if b.option.docDomain == "" {
			swagger.SetHost(fmt.Sprintf("http://%s:%d", b.option.intranetIP, b.option.listenPort))
		} else {
			_log.Debug("域名:" + b.option.docDomain)
			swagger.SetHost(b.option.docDomain)
		}

		swagger.SetBasePath("")
		swagger.SetSchemes(true, true)
		swagger.SetInfo(swagger.Info{
			Description: b.option.docName,
			Version:     b.option.docVer,
			Title:       b.option.docDesc,
		})

	}
	b.baseGroup = b.engine.Group(b.option.apiBasePath)
	b.Register(b.baseGroup, cList...)
}

func (b *KApi) Run() {
	b.genRouterCode()
	b.engine.GET("/version", func(c *gin.Context) {
		c.String(200, "ServerTime:%36s\nAPI_Version_Time:%30s", time.Now().Format(time.RFC3339), time.Unix(_genInfo.Tm, 0).Format(time.RFC3339))
	})
	if b.option.needDoc {
		swaggerUrl := fmt.Sprintf("http://%s:%d/swagger/index.html", b.option.intranetIP, b.option.listenPort)
		if b.option.docDomain != "" {
			swaggerUrl = fmt.Sprintf("%s/swagger/index.html", b.option.docDomain)
		}
		_log.Infoln("文档地址:", swaggerUrl)
		if b.option.redirectToDocWhenAccessRoot {
			b.engine.Any("/", func(c *gin.Context) {
				c.Redirect(301, "/swagger/")
			})
		}

		// 将解析的文档内容直接通过接口返回, 省去了写到文件的步骤
		b.engine.GET("/swagger.json", func(c *gin.Context) {
			c.JSON(200, b.doc.Client)
		})
		b.engine.GET("/swagger/*any", func(c *gin.Context) {
			c.FileFromFS(c.Request.URL.Path, http.FS(swaggerFS))
		})

		if b.option.openDocInBrowser {
			err := internal.OpenBrowser(swaggerUrl)
			if err != nil {
				_log.Error(err)
			}
		}
	}
	_log.Infoln(fmt.Sprintf("服务启动 http://%s:%d", b.option.intranetIP, b.option.listenPort))
	err := b.engine.Run(fmt.Sprintf(":%d", b.option.listenPort))
	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e1, ok := e.Err.(*os.SyscallError); ok {
				if e.Op == "listen" && e1.Syscall == "bind" {
					_log.Fatalf("服务启动失败, 监听 :%d 失败, 请检查端口是否被占用", e.Addr.(*net.TCPAddr).Port)
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
func (b *KApi) Register(router gin.IRoutes, cList ...interface{}) bool {
	//开发时 生成路由注册文件
	if b.option.isDebug {
		b.tryGenRegister(router, cList...)
	} else {
		// 为处理一种情况: 没有run过, 直接build的情况 不会生成gen_router.go 此时路由信息需要从代码中读取
		// 这里即使使用 ldflags="-w -s" 也可以正常读取
		if len(_genInfo.List) <= 0 {
			b.tryGenRegister(router, cList...)
		}
	}

	return b.register(router, cList...)
}

// genRouterCode 生成gen_router.go
func (b *KApi) genRouterCode() {
	if !b.option.isDebug {
		return
	}
	_, modFile, _ := ast.GetModuleInfo(2)
	genOutPut(b.option.outPath, modFile)
}

var _ctlsList = make([]interface{}, 0)

func RegisterController(c interface{}) {
	_ctlsList = append(_ctlsList, c)
}
