package internal

import (
	"bytes"
	"encoding/gob"
	"github.com/go-openapi/spec"
	"sync"
	"time"
)

// RouteItem store the comment from controller's method.
// fields should be exported for gob encoding and decoding
type RouteItem struct {
	Key         string //may be [controller name] + / + [method name]. unique
	RouterPath  string
	Summary     string
	Description string
	Method      string //HTTP METHOD
}

type genInfo struct {
	Routes    []RouteItem
	Swagger   *spec.Swagger
	Timestamp int64 //timestamp of this
}

type RouteInfo struct {
	once    sync.Once
	mu      sync.Mutex
	genInfo *genInfo
}

func NewRouteInfo() *RouteInfo {
	ri := &RouteInfo{}
	ri.genInfo = &genInfo{
		Routes:    []RouteItem{},
		Swagger:   &spec.Swagger{},
		Timestamp: time.Now().Unix(),
	}
	ri.load()
	return ri
}

func init() {
	gob.Register(map[string]string{})
}

// AddFunc add one method to method comments
func (ri *RouteInfo) AddFunc(handlerFuncName, routerPath string, method string) {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	ri.genInfo.Routes = append(ri.genInfo.Routes, RouteItem{
		Key:        handlerFuncName,
		RouterPath: routerPath,
		Method:     method,
	})
}
func (ri *RouteInfo) GetGenInfo() *genInfo {
	return ri.genInfo
}

// SetApiBody store swagger json spec
//
//	@param api
func (ri *RouteInfo) SetApiBody(api *spec.Swagger) {
	ri.genInfo.Swagger = api
}

// WriteOut write router info to gen.gob
func (ri *RouteInfo) WriteOut() {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	ri.genInfo.Timestamp = time.Now().Unix()
	err := encoder.Encode(ri.genInfo)
	if err != nil {
		//kapi.Errorf("%s", err)
		return
	}
	WriteFile("gen.gob", buf.Bytes(), true)
}

// GetRouteItems get router info of method comments
//
//	@return map[string][]RouteItem
func (ri *RouteInfo) GetRouteItems() map[string][]RouteItem {
	ri.mu.Lock()
	defer ri.mu.Unlock()

	mp := make(map[string][]RouteItem, len(ri.genInfo.Routes))
	for _, v := range ri.genInfo.Routes {
		tmp := v
		mp[tmp.Key] = append(mp[tmp.Key], tmp)
	}
	return mp
}
func (ri *RouteInfo) Clean() {
	ri.genInfo = &genInfo{
		Routes:    []RouteItem{},
		Swagger:   &spec.Swagger{},
		Timestamp: time.Now().Unix(),
	}
}
func (ri *RouteInfo) load() {
	if FileIsExist("gen.gob") {
		data := ReadFile("gen.gob")
		var buf = bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		_ = dec.Decode(ri.genInfo)
	}
}
