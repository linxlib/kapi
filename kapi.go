package kapi

import (
	"embed"
	"fmt"
	"gitee.com/kirile/kapi/ast"
	"gitee.com/kirile/kapi/doc/swagger"
	"gitee.com/kirile/kapi/tools"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/linxlib/logs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
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
var _log logs.FieldLogger = &logs.Logger{
	Out:   os.Stderr,
	Hooks: make(logs.LevelHooks),
	Formatter: &logs.TextFormatter{
		ForceColors:      true,
		FullTimestamp:    true,
		TimestampFormat:  "01-02 15:04:05.999",
		QuoteEmptyFields: false,
		CallerPrettyfier: nil,
		HideLevelText:    true,
	},
	ReportCaller: false,
	Level:        logs.TraceLevel,
	//ReportFunc:   false,
}

// KApi base struct
type KApi struct {
	apiFun  ApiFunc
	apiType reflect.Type

	beforeAfter Interceptor //拦截器

	recoverErrorFunc RecoverErrorFunc
	engine           *gin.Engine
	doc              *swagger.DocSwagger
	outPath          string // output path.输出目录
	isOutDoc         bool
	docName          string
	isOpenDoc        bool
	isDev            bool // if is development
	port             int
	domain           string
	intranetIP       string
}

// Option overrides behavior of Connect.
type Option interface {
	apply(*KApi)
}

type optionFunc func(*KApi)

func (f optionFunc) apply(o *KApi) {
	f(o)
}

// WithOutPath gen_router.go 路由注册文件输出目录 默认为 ./routers/gen_router.go
func WithOutPath(path string) Option {

	return optionFunc(func(o *KApi) {
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		o.outPath = path
	})
}

// WithCtx use custom context.设置自定义context
func WithCtx(middleware ApiFunc) Option {
	return optionFunc(func(o *KApi) {
		o.Model(middleware)
	})
}

// WithDebug 设置模式(默认dev模式) 当非调试模式时, 会一并将swagger文档功能也关闭
func WithDebug(b bool) Option {
	return optionFunc(func(o *KApi) {
		o.Dev(b)
		if !b {
			o.isOutDoc = false
			o.domain = ""
		}
	})
}

// WithImportFile 添加自定义import文件列表
func WithImportFile(k, v string) Option {
	return optionFunc(func(o *KApi) {
		ast.AddImportFile(k, v)
	})
}

// OutputDoc 是否输出文档
func OutputDoc(name ...string) Option {
	return optionFunc(func(o *KApi) {
		o.isOutDoc = true
		if len(name) > 0 {
			o.docName = name[0]
		}
	})
}

// OpenDoc 打开浏览器浏览文档
func OpenDoc() Option {
	return optionFunc(func(o *KApi) {
		o.isOpenDoc = true
	})
}
func WithDomain(domain string) Option {
	return optionFunc(func(o *KApi) {
		o.domain = domain
	})
}
func Port(port int) Option {
	return optionFunc(func(o *KApi) {
		o.port = port
	})
}

// WithBeforeAfter 设置对象调用前后执行中间件
func WithBeforeAfter(beforeAfter Interceptor) Option {
	return optionFunc(func(o *KApi) {
		o.beforeAfter = beforeAfter
	})
}

var defaultLogFormatter = func(param gin.LogFormatterParams) string {
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
		ByteCountSI(int64(param.BodySize)),
		param.ErrorMessage,
	)
}

//ByteCountSI 字节数转带单位
func ByteCountSI(b int64) string {
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

// New 新
func New(opts ...Option) *KApi {
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

	b := new(KApi) // default option
	b.Model(NewAPIFunc)
	b.Dev(true)
	b.intranetIP = tools.GetIntranetIp()

	b.engine = gin.New()
	b.engine.Use(gin.LoggerWithFormatter(defaultLogFormatter))
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Token"}
	b.engine.Use(cors.New(corsConfig))
	b.docName = "Test"
	b.port = 8080

	for _, o := range opts {
		o.apply(b)
	}
	_log.Infof("localIP:%s docName:%s port:%d", b.intranetIP, b.docName, b.port)
	b.SetRecover(func(err interface{}) {
		switch err {
		case KAPIEXIT:
			return
		default:
			_log.Error(err)
		}
	})
	return b
}

// Dev set build is development
func (b *KApi) Dev(isDev bool) {
	b.isDev = isDev
	gin.SetMode(gin.ReleaseMode)
}

func (b *KApi) RegisterRouter(f func(b *KApi, r *gin.Engine)) {
	_log.Debug("注册路由..")
	if b.isOutDoc {
		if b.domain == "" {
			swagger.SetHost(fmt.Sprintf("http://%s:%d", b.intranetIP, b.port))
		} else {
			_log.Debug("域名:" + b.domain)
			swagger.SetHost(b.domain)
		}

		swagger.SetBasePath("")
		swagger.SetSchemes(true, true)
		swagger.SetInfo(swagger.Info{
			Description: b.docName,
			Version:     "v1",
			Title:       b.docName,
		})
		if b.doc == nil {
			b.doc = swagger.NewDoc()
		}

	} else {
		if b.doc == nil {
			b.doc = swagger.NewDoc()
		}
	}

	f(b, b.engine)
}

//go:embed swagger/*
var swaggerFS embed.FS

func (b *KApi) Run() {
	b.genRouterCode()
	b.engine.GET("/version", func(c *gin.Context) {
		c.String(200, "ServerTime:%36s\nAPI_Version_Time:%30s", time.Now().Format(time.RFC3339), time.Unix(_genInfo.Tm, 0).Format(time.RFC3339))
	})
	if b.isOutDoc {
		//saduygas
		//b.engine.StaticFile("/swagger.json", "./swagger/swagger.json")
		//url := func(c *ginSwagger.Config) {}
		swaggerUrl := ""
		if b.domain != "" {
			swaggerUrl = fmt.Sprintf("%s/swagger/index.html", b.domain)
		} else {
			swaggerUrl = fmt.Sprintf("http://%s:%d/swagger/index.html", b.intranetIP, b.port)
		}
		_log.Infoln("文档地址:", swaggerUrl)
		b.engine.Any("/", func(c *gin.Context) {
			c.Redirect(301, "/swagger/")
		})
		// 将解析的文档内容直接通过接口返回, 省去了写到文件的步骤
		b.engine.GET("/swagger.json", func(c *gin.Context) {
			c.JSON(200, b.doc.Client)
		})
		b.engine.GET("/swagger/*any", func(c *gin.Context) {
			c.FileFromFS(c.Request.URL.Path, http.FS(swaggerFS))
		})
		//b.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
		if b.isOpenDoc {
			if b.domain != "" {
				err := openBrowser(swaggerUrl)
				if err != nil {
					_log.Error(err)
				}
			} else {
				err := openBrowser(swaggerUrl)
				if err != nil {
					_log.Error(err)
				}
			}

		}
	}
	_log.Infoln(fmt.Sprintf("服务启动 http://%s:%d", b.intranetIP, b.port))
	err := b.engine.Run(fmt.Sprintf(":%d", b.port))
	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e1, ok := e.Err.(*os.SyscallError); ok {
				if e.Op == "listen" && e1.Syscall == "bind" {
					_log.Fatalf("服务启动失败, 监听 :%d 失败, 请检查端口是否被占用", e.Addr.(*net.TCPAddr).Port)
				}
			}

		}
		b.recoverErrorFunc(err)
	}
}

// Open calls the OS default program for uri
func openBrowser(uri string) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	cmd := exec.Command("cmd", "/c", "start", uri)
	return cmd.Start()
}

// SetRecover set recover err call
func (b *KApi) SetRecover(f func(interface{})) {
	b.recoverErrorFunc = f
}

// OutDoc set if out doc. 设置是否输出接口文档
func (b *KApi) OutDoc(isOutDoc bool) {
	b.isOutDoc = isOutDoc
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
	b.apiType = rt

	return b
}

// Register Registered by struct object,[prepath + objname.]
func (b *KApi) Register(router gin.IRoutes, cList ...interface{}) bool {
	//开发时 生成路由注册文件
	if b.isDev {
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

// RegisterHandlerFunc 直接注册某个路由方法
//func (b *KApi) RegisterHandlerFunc(router gin.IRoutes, httpMethod []string, relativePath string, handlerFuncs ...interface{}) error {
//	list := make([]gin.HandlerFunc, 0, len(handlerFuncs))
//	for _, call := range handlerFuncs {
//		list = append(list, b.HandlerFunc(call))
//	}
//
//	for _, v := range httpMethod {
//		// method := strings.ToUpper(v)
//		// switch method{
//		// case "ANY":
//		// 	router.Any(relativePath,list...)
//		// default:
//		// 	router.Handle(method,relativePath,list...)
//		// }
//		// or
//		switch strings.ToUpper(v) {
//		case "POST":
//			router.POST(relativePath, list...)
//		case "GET":
//			router.GET(relativePath, list...)
//		case "DELETE":
//			router.DELETE(relativePath, list...)
//		case "PATCH":
//			router.PATCH(relativePath, list...)
//		case "PUT":
//			router.PUT(relativePath, list...)
//		case "OPTIONS":
//			router.OPTIONS(relativePath, list...)
//		case "HEAD":
//			router.HEAD(relativePath, list...)
//		case "ANY":
//			router.Any(relativePath, list...)
//		default:
//			return errors.Errorf("请求方式:[%v] 不支持", httpMethod)
//		}
//	}
//
//	return nil
//}

// HandlerFunc 转换路由方法到gin的路由方法格式
//func (b *KApi) HandlerFunc(handlerFunc interface{}) gin.HandlerFunc { // 获取并过滤要绑定的参数
//	typ := reflect.ValueOf(handlerFunc).Type() //反射获取路由方法的参数信息
//	// 当路由方法只有一个上下文参数时
//	if typ.NumIn() == 1 {
//		ctxType := typ.In(0)
//		// go-gin 默认的上下文参数
//		if ctxType == reflect.TypeOf(&gin.Context{}) {
//			return handlerFunc.(func(*gin.Context))
//		}
//
//		// 自定义的context
//		if ctxType == b.apiType {
//			method := reflect.ValueOf(handlerFunc) //获取方法指针
//			return func(c *gin.Context) {
//				//方法调用中出现panic，则调用 recoverErrorFunc 进行处理
//				defer func() {
//					if err := recover(); err != nil {
//						b.recoverErrorFunc(err)
//					}
//				}()
//				method.Call([]reflect.Value{reflect.ValueOf(b.apiFun(c))})
//			}
//		}
//
//		panic("方法 " + runtime.FuncForPC(reflect.ValueOf(handlerFunc).Pointer()).Name() + " 不支持!")
//	}
//
//	// 自定义的context类型, 带request 请求参数
//	call, err := b.changeToGinHandler(reflect.ValueOf(handlerFunc))
//	if err != nil { // Direct reporting error.
//		panic(err)
//	}
//
//	return call
//}

// CheckHandlerFunc Judge whether to match rules
//func (b *KApi) CheckHandlerFunc(handlerFunc interface{}) (int, bool) { // 判断是否匹配规则,返回参数个数
//	typ := reflect.ValueOf(handlerFunc).Type()
//	return b.checkHandlerFunc(typ, false)
//}

// genRouterCode 生成gen_router.go
func (b *KApi) genRouterCode() {
	if !b.isDev {
		return
	}
	_, modFile, _ := ast.GetModuleInfo(2)
	genOutPut(b.outPath, modFile)
}

// genSwagger 生成swagger.json
func (b *KApi) genSwagger() string {
	if !b.isDev {
		return ""
	}
	//_, modFile, isFind := myast.GetModuleInfo(2)
	//if !isFind {
	//	return ""
	//}
	return b.doc.GetAPIString()
	//_log.Debugf("生成 %s/swagger/swagger.json", modFile)
	//tools.WriteFile(modFile+"/swagger/swagger.json", []string{jsonsrc}, true)
	//return true
}
