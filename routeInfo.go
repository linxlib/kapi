package kapi

import (
	"bytes"
	"encoding/gob"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/swagger"
	"sync"
	"time"
)

// MethodComment store the comment from controller's method.
// fields should be exported for gob encoding and decoding
type MethodComment struct {
	Key          string //may be [controller name] + / + [method name]. unique
	RouterPath   string
	IsDeprecated bool   // will show deprecated in swagger ui
	ResultType   string //then response type name. eg. model.User or []model.User
	Summary      string
	Description  string
	Method       string //HTTP METHOD
	TokenHeader  string
}

type genInfo struct {
	Methods []MethodComment
	ApiBody swagger.APIBody
	Tm      int64 //timestamp of this
}

type RouteInfo struct {
	once    sync.Once
	mu      sync.Mutex
	genInfo *genInfo
}

var routeInfo *RouteInfo

func init() {
	//swagger.APIBody contains map[string]string , need to be registered
	gob.Register(map[string]string{})
	routeInfo = new(RouteInfo)
	routeInfo.genInfo = &genInfo{
		Methods: []MethodComment{},
		ApiBody: swagger.APIBody{},
		Tm:      time.Now().Unix(),
	}
	//load from previously generated gen.gob file
	//on production environment, we don't have codes and cannot analyze comments
	if internal.FileIsExist("gen.gob") {
		data := internal.ReadFile("gen.gob")
		var buf = bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		_ = dec.Decode(routeInfo.genInfo)
	}
}

// AddFunc add one method to method comments
func (ri *RouteInfo) AddFunc(handlerFuncName, routerPath string, method string) {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	// when AddFunc called first time, init these fields (only be called in debug mode)
	// so that we use newly generated route info instead of gen.gob file
	ri.once.Do(func() {
		ri.genInfo.Tm = time.Now().Unix()
		ri.genInfo.Methods = []MethodComment{}
		ri.genInfo.ApiBody = swagger.APIBody{}
	})
	ri.genInfo.Methods = append(ri.genInfo.Methods, MethodComment{
		Key:        handlerFuncName,
		RouterPath: routerPath,
		Method:     method,
	})
}

//SetApiBody store swagger json spec
//  @param api
//
func (ri *RouteInfo) SetApiBody(api swagger.APIBody) {
	ri.genInfo.ApiBody = api
}

//writeOut write router info to gen.gob
//
func (ri *RouteInfo) writeOut() {
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

//getInfo get router info of method comments
//
//  @return map[string][]MethodComment
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
