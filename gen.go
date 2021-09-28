package kapi

import (
	"bytes"
	"gitee.com/kirile/kapi/internal"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/template"
	"time"
)

var _mu sync.Mutex // protects the serviceMap
var _once sync.Once
var _genInfo genInfo

// AddGenOne add one to base case
func AddGenOne(handFunName, routerPath string, methods []string) {
	_mu.Lock()
	defer _mu.Unlock()
	_genInfo.List = append(_genInfo.List, genRouterInfo{
		HandFunName: handFunName,
		GenComment: genComment{
			RouterPath: routerPath,
			Methods:    methods,
		},
	})
}

// SetVersion use timestamp to replace version
func SetVersion(tm int64) {
	_mu.Lock()
	defer _mu.Unlock()
	_genInfo.Tm = tm
}

func checkOnceAdd(handFunName, routerPath string, methods []string) {
	_once.Do(func() {
		_mu.Lock()
		defer _mu.Unlock()
		_genInfo.Tm = time.Now().Unix()
		_genInfo.List = []genRouterInfo{} // reset
	})

	AddGenOne(handFunName, routerPath, methods)
}

// getStringList format string
func getStringList(list []string) string {
	return `"` + strings.Join(list, `","`) + `"`
}

func genOutPut(outDir, modFile string) {

	_mu.Lock()
	defer _mu.Unlock()

	genCode(outDir, modFile) // gen .go file
	//_genInfo.Tm = time.Now().Unix()
	//_data, _ := tools.Encode(&_genInfo) // gob serialize 序列化
	//_path := modFile+getRouter
	//tools.BuildDir(_path)
	//f, err := os.Create(_path)
	//if err != nil {
	//	return
	//}
	//defer f.Close()
	//f.Write(_data)
}

func genCode(outDir, modFile string) bool {
	_genInfo.Tm = time.Now().Unix()
	if len(outDir) == 0 {
		outDir = modFile + "/routers/"
	}
	pkgName := getPkgName(outDir)
	data := struct {
		genInfo
		PkgName string
		T       string
	}{
		genInfo: _genInfo,
		PkgName: pkgName,
		T:       time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("gen_out").Funcs(template.FuncMap{"getStringList": getStringList}).Parse(genTemp)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	tmpl.Execute(&buf, data)
	f, err := os.Create(outDir + "gen_router.go")
	if err != nil {
		return false
	}
	defer f.Close()
	f.Write(buf.Bytes())
	_log.Debug("生成 ", outDir, "gen_router.go")
	// format
	exec.Command("gofmt", "-l", "-w", outDir).Output()
	return true
}

func getPkgName(dir string) string {
	dir = strings.Replace(dir, "\\", "/", -1)
	dir = strings.TrimRight(dir, "/")

	var pkgName string
	list := strings.Split(dir, "/")
	if len(list) > 0 {
		pkgName = list[len(list)-1]
	}

	if len(pkgName) == 0 || pkgName == "." {
		list = strings.Split(internal.GetCurrentDirectory(), "/")
		if len(list) > 0 {
			pkgName = list[len(list)-1]
		}
	}

	return pkgName
}

// 获取路由注册信息
func getInfo() map[string][]genRouterInfo {
	_mu.Lock()
	defer _mu.Unlock()

	mp := make(map[string][]genRouterInfo, len(_genInfo.List))
	for _, v := range _genInfo.List {
		tmp := v
		mp[tmp.HandFunName] = append(mp[tmp.HandFunName], tmp)
	}
	return mp
}

func buildRelativePath(prepath, routerPath string) string {
	if strings.HasSuffix(prepath, "/") {
		if strings.HasPrefix(routerPath, "/") {
			return prepath + strings.TrimPrefix(routerPath, "/")
		}
		return prepath + routerPath
	}

	if strings.HasPrefix(routerPath, "/") {
		return prepath + routerPath
	}

	return prepath + "/" + routerPath
}
