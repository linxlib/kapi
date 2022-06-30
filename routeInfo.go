package kapi

import (
	"bytes"
	"encoding/gob"
	"github.com/linxlib/kapi/doc/swagger"
	"github.com/linxlib/kapi/internal"
	"sync"
	"time"
)

// store the comment for the controller method. 生成注解路由
type MethodComment struct {
	Key          string
	RouterPath   string
	IsDeprecated bool
	ResultType   string
	Summary      string //方法说明
	Description  string // 方法注释
	Methods      []string
	TokenHeader  string
}

type genInfo struct {
	Methods []MethodComment
	ApiBody swagger.APIBody
	Tm      int64
}

var routeInfo *RouteInfo

type RouteInfo struct {
	once    sync.Once
	mu      sync.Mutex
	genInfo *genInfo
}

func init() {
	gob.Register(map[string]string{})
	routeInfo = new(RouteInfo)
	routeInfo.genInfo = &genInfo{
		Methods: []MethodComment{},
		ApiBody: swagger.APIBody{},
		Tm:      time.Now().Unix(),
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
		ri.genInfo.Methods = []MethodComment{}
		ri.genInfo.ApiBody = swagger.APIBody{}
	})
	ri.genInfo.Methods = append(ri.genInfo.Methods, MethodComment{
		Key:        handFunName,
		RouterPath: routerPath,
		Methods:    methods,
	})
}

func (ri *RouteInfo) checkOnceAdd(handFunName, routerPath string, methods []string) {
	ri.AddFunc(handFunName, routerPath, methods)

}
func (ri *RouteInfo) SetApiBody(api swagger.APIBody) {
	ri.genInfo.ApiBody = api
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

func (ri *RouteInfo) getApiBody() swagger.APIBody {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	return ri.genInfo.ApiBody
}

// 获取路由注册信息
func (ri *RouteInfo) getInfo() map[string][]MethodComment {
	ri.mu.Lock()
	defer ri.mu.Unlock()

	mp := make(map[string][]MethodComment, len(ri.genInfo.Methods))
	for _, v := range ri.genInfo.Methods {
		tmp := v

		mp[tmp.Key] = append(mp[tmp.Key], tmp)
	}
	return mp
}
