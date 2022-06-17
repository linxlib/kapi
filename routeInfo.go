package kapi

import (
	"bytes"
	"encoding/gob"
	"github.com/linxlib/kapi/internal"
	"sync"
	"time"
)

var routeInfo *RouteInfo

type RouteInfo struct {
	once    sync.Once
	mu      sync.Mutex
	genInfo *genInfo
}

func init() {
	routeInfo = new(RouteInfo)
	routeInfo.genInfo = &genInfo{
		List: []genRouterInfo{},
		Tm:   time.Now().Unix(),
	}
	if internal.FileIsExist("gen.gob") {
		data := internal.ReadFile("gen.gob")
		var buf = bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		dec.Decode(routeInfo.genInfo)
	}
}

// AddFunc add one to base case
func (ri *RouteInfo) AddFunc(handFunName, routerPath string, methods []string) {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	ri.once.Do(func() {
		ri.genInfo.Tm = time.Now().Unix()
		ri.genInfo.List = []genRouterInfo{}
	})
	ri.genInfo.List = append(ri.genInfo.List, genRouterInfo{
		HandFunName: handFunName,
		GenComment: genComment{
			RouterPath: routerPath,
			Methods:    methods,
		},
	})
}

func (ri *RouteInfo) checkOnceAdd(handFunName, routerPath string, methods []string) {
	ri.AddFunc(handFunName, routerPath, methods)
}

func (ri *RouteInfo) genOutPut() {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	ri.genInfo.Tm = time.Now().Unix()
	err := encoder.Encode(ri.genInfo)
	if err != nil {
		internal.Log.Error(err)
		return
	}
	internal.WriteFile("gen.gob", buf.Bytes(), true)
}

// 获取路由注册信息
func (ri *RouteInfo) getInfo() map[string][]genRouterInfo {
	ri.mu.Lock()
	defer ri.mu.Unlock()

	mp := make(map[string][]genRouterInfo, len(ri.genInfo.List))
	for _, v := range ri.genInfo.List {
		tmp := v

		mp[tmp.HandFunName] = append(mp[tmp.HandFunName], tmp)
	}
	return mp
}
