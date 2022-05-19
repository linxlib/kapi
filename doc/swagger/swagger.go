package swagger

import (
	"github.com/linxlib/kapi/internal"
	"strings"
)

// DocSwagger ...
type DocSwagger struct {
	Client *APIBody
}

// NewDoc 新建一个swagger doc
func NewDoc(host string, info Info, basePath string, schemes []string) *DocSwagger {
	doc := &DocSwagger{}
	doc.Client = &APIBody{
		Head: Head{Swagger: version},
		Info: info,
		Host: host,
		SecurityDefinitions: &SecurityDefinitions{
			Type: "apiKey",
			Name: "Authorization",
			In:   "header",
		},
		BasePath:     basePath,
		Schemes:      schemes,
		ExternalDocs: nil,
		Definitions:  map[string]Definition{},
		Tags:         []Tag{},
	}
	doc.Client.Paths = make(map[string]map[string]Param)
	return doc
}

// AddTag add tag (排他)
func (doc *DocSwagger) AddTag(tag Tag) {
	for _, v := range doc.Client.Tags {
		if v.Name == tag.Name { // find it
			return
		}
	}

	doc.Client.Tags = append(doc.Client.Tags, tag)
}

// AddDefinitions 添加 结构体定义
func (doc *DocSwagger) AddDefinitions(key string, def Definition) {
	doc.Client.Definitions[key] = def
}

// AddPatch ... API 路径 paths 和操作在 API 规范的全局部分定义
func (doc *DocSwagger) AddPatch(url string, p Param, methods ...string) {
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}

	if doc.Client.Paths[url] == nil {
		doc.Client.Paths[url] = make(map[string]Param)
	}
	if len(p.Consumes) == 0 {
		p.Consumes = reqCtxType
	}
	if len(p.Produces) == 0 {
		p.Produces = respCtxType
	}
	if p.Responses == nil {
		p.Responses = map[string]Resp{
			"500": {Description: "操作异常"},
			"400": {Description: "参数错误"},
			"401": {Description: "鉴权失败"},
			"403": {Description: "权限问题"},
			"404": {Description: "资源未找到"},
		}
	} else {
		p.Responses["500"] = Resp{Description: "发生异常"}
		p.Responses["400"] = Resp{Description: "参数错误"}
		p.Responses["401"] = Resp{Description: "鉴权失败"}
		p.Responses["403"] = Resp{Description: "权限问题"}
		p.Responses["404"] = Resp{Description: "资源未找到"}
	}
	for _, v := range methods {
		doc.Client.Paths[url][strings.ToLower(v)] = p
	}
}

// GetAPIString 获取返回数据
func (doc *DocSwagger) GetAPIString() string {
	return internal.MarshalToJson(doc.Client, true)
}

var kvType = map[string]string{ // array, boolean, integer, number, object, string
	"int":     "integer",
	"uint":    "integer",
	"byte":    "integer",
	"rune":    "integer",
	"int8":    "integer",
	"int16":   "integer",
	"int32":   "integer",
	"int64":   "integer",
	"uint8":   "integer",
	"uint16":  "integer",
	"uint32":  "integer",
	"uint64":  "integer",
	"uintptr": "integer",
	"float32": "integer",
	"float64": "integer",
	"bool":    "boolean",
	"map":     "object",
	"string":  "string",
	"Time":    "string"}

var kvFormat = map[string]string{}

// GetKvType 获取类型转换
func GetKvType(k string, isArray, isType bool) string {
	if isArray {
		if isType {
			return "array"
		}
		return "array"
	}

	if isType {
		if _, ok := kvType[k]; ok {
			return kvType[k]
		}
		return "object"
	}
	if _, ok := kvFormat[k]; ok {
		return kvFormat[k]
	}
	return k
}
