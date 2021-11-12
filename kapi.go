package kapi

import (
	"embed"
	"encoding/json"
	"fmt"
	"gitee.com/kirile/kapi/ast"
	"gitee.com/kirile/kapi/doc/swagger"
	"gitee.com/kirile/kapi/internal"
	"gitee.com/kirile/kapi/internal/cors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
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

	gin.SetMode(gin.ReleaseMode) //we don't need gin's debug output
	b.engine = gin.New()
	b.engine.Use(gin.LoggerWithFormatter(b.option.ginLoggerFormatter))
	b.engine.Use(cors.New(b.option.corsConfig))

	_log.Infof("localIP:%s docName:%s port:%d", b.option.intranetIP, b.option.docName, b.option.listenPort)

	return b
}

func (b *KApi) RegisterRouter(cList ...interface{}) {
	_log.Debug("注册路由..")

	b.handleSwaggerBase()
	b.baseGroup = b.engine.Group(b.option.apiBasePath)
	b.doRegister(b.baseGroup, cList...)
	b.handleStatic()
}

func (b *KApi) handleSwaggerBase() {
	schemes := []string{"http"}
	host := "localhost"
	basePath := ""
	info := swagger.Info{
		Description: "kapi api server",
		Version:     "1.0.0",
		Title:       "KAPI",
	}
	if b.option.needDoc {
		if b.option.docDomain == "" {
			schemes = []string{"http"}
			host = fmt.Sprintf("%s:%d", b.option.intranetIP, b.option.listenPort)
		} else {
			_log.Debug("域名:" + b.option.docDomain)
			if strings.HasPrefix(b.option.docDomain, "https") {
				schemes = []string{"https", "http"}
			} else {
				schemes = []string{"http", "https"}
			}
			host = strings.TrimPrefix(b.option.docDomain, "http://")
			host = strings.TrimPrefix(host, "https://")
		}
		info = swagger.Info{
			Description: b.option.docName,
			Version:     b.option.docVer,
			Title:       b.option.docDesc,
		}
	}
	b.doc = swagger.NewDoc(host, info, basePath, schemes)
}

func (b *KApi) handleStatic() {
	if b.option.staticDir != "" {
		b.engine.Static(b.option.staticDir, b.option.staticDir)
	}
}

func (b *KApi) handleSwaggerDoc() {
	if b.option.needDoc {
		swaggerUrl := fmt.Sprintf("http://%s:%d/swagger/index.html", b.option.intranetIP, b.option.listenPort)
		if b.option.docDomain != "" {
			swaggerUrl = fmt.Sprintf("%s/swagger/index.html", b.option.docDomain)
		}
		_log.Infoln("文档地址:", swaggerUrl)
		if b.option.redirectToDocWhenAccessRoot {
			b.engine.Any("/", func(c *gin.Context) {
				c.Redirect(301, fmt.Sprintf("%s/swagger/", b.option.docDomain))
			})
		}

		// 将解析的文档内容直接通过接口返回, 省去了写到文件的步骤
		b.engine.GET("/swagger.json", func(c *gin.Context) {
			if !internal.CheckFileIsExist("swagger.json") {
				_log.Warningln("swagger.json未找到, 请放置到程序目录下")
				c.JSON(404, "swagger.json not found")
				return
			}
			bs, _ := ioutil.ReadFile("swagger.json")
			c.String(200, string(bs))
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
}

func (b *KApi) Run() {
	b.genRouterCode()
	b.genDoc()
	b.handleSwaggerDoc()

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
func (b *KApi) doRegister(router gin.IRoutes, cList ...interface{}) bool {
	//开发时 生成路由注册文件 gen.gob
	if b.option.isDebug {
		b.tryGenRegister(router, cList...)
	}

	return b.register(router, cList...)
}

// genRouterCode 生成gen.gob
func (b *KApi) genRouterCode() {
	if !b.option.isDebug {
		return
	}
	_, modFile, _ := ast.GetModuleInfo(2)
	genOutPut(b.option.outPath, modFile)
}

// genDoc 生成swagger.json
func (b *KApi) genDoc() {
	if !b.option.isDebug || !b.option.needDoc {
		return
	}
	bs, _ := json.Marshal(b.doc.Client)

	_, modFile, _ := ast.GetModuleInfo(2)
	outDir := b.option.outPath
	if len(outDir) == 0 {
		outDir = modFile + "/"
	}
	if outDir[len(outDir)-1] != '/' {
		outDir += "/"
	}
	err := ioutil.WriteFile(outDir+"swagger.json", bs, os.ModePerm)
	if err != nil {
		_log.Errorln("写出 swagger.json 失败:", err.Error())
		return
	}

}

var _ctlsList = make([]interface{}, 0)

func RegisterController(c interface{}) {
	_ctlsList = append(_ctlsList, c)
}
