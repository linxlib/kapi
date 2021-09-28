package swagger

import (
	"gitee.com/kirile/kapi/internal"
	"strings"
)

// DocSwagger ...
type DocSwagger struct {
	Client *APIBody
}

// NewDoc 新建一个swagger doc
func NewDoc() *DocSwagger {
	doc := &DocSwagger{}
	doc.Client = &APIBody{
		Head:     Head{Swagger: version},
		Info:     info,
		Host:     host,
		BasePath: basePath,
		Schemes:  schemes,
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
	if doc.Client.Definitions == nil {
		doc.Client.Definitions = make(map[string]Definition)
	}
	doc.Client.Definitions[key] = def
}

// AddPatch ... API 路径 paths 和操作在 API 规范的全局部分定义
func (doc *DocSwagger) AddPatch(url string, p Param, metheds ...string) {
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
			"500": {Description: "操作失败"},
			"400": {Description: "参数错误"},
			"401": {Description: "鉴权失败"},
		}
	} else {
		p.Responses["500"] = Resp{Description: "发生错误"}
		p.Responses["400"] = Resp{Description: "参数错误 失败"}
		p.Responses["401"] = Resp{Description: "鉴权失败"}
	}

	for _, v := range metheds {
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
