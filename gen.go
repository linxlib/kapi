package kapi

import (
	"bytes"
	"encoding/gob"
	"gitee.com/kirile/kapi/internal"
	"strings"
	"sync"
	"time"
)

var _mu sync.Mutex // protects the serviceMap
var _once sync.Once
var _genInfo genInfo

func init() {
	if internal.CheckFileIsExist("gen.gob") {
		data := internal.ReadFile("gen.gob")
		var buf = bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		dec.Decode(&_genInfo)
	}
}

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

func checkOnceAdd(handFunName, routerPath string, methods []string) {
	_once.Do(func() {
		_mu.Lock()
		defer _mu.Unlock()
		_genInfo.Tm = time.Now().Unix()
		_genInfo.List = []genRouterInfo{} // reset
	})

	AddGenOne(handFunName, routerPath, methods)
}

func genOutPut() {
	_mu.Lock()
	defer _mu.Unlock()
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	_genInfo.Tm = time.Now().Unix()
	err := encoder.Encode(_genInfo)
	if err != nil {
		_log.Error(err)
		return
	}
	internal.WriteFile("gen.gob", buf.Bytes(), true)
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
