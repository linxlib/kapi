package swagger

import (
	"github.com/linxlib/kapi/doc"
	"github.com/linxlib/kapi/internal"
	"strings"
)

// DocSwagger ...
type DocSwagger struct {
	Client *APIBody
}

// NewDoc 新建一个swagger doc
func NewDoc(host string, info Info, basePath string, schemes []string) *DocSwagger {
	ds := &DocSwagger{}
	ds.Client = &APIBody{
		Head:         Head{Swagger: version},
		Info:         info,
		Host:         host,
		BasePath:     basePath,
		Schemes:      schemes,
		ExternalDocs: nil,
		Definitions:  map[string]Definition{},
		Tags:         []Tag{},
	}
	ds.Client.Paths = make(map[string]map[string]Param)
	return ds
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

func (doc *DocSwagger) AddPatch2(url string, p Param, method string) {
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
	doc.Client.Paths[url][strings.ToLower(method)] = p
}

// GetAPIString 获取返回数据
func (doc *DocSwagger) GetAPIString() string {
	return internal.MarshalToJson(doc.Client, true)
}

func (doc *DocSwagger) SetDefinition(m *doc.Model, si *doc.StructInfo) string {
	if si == nil {
		return ""
	}
	var def = Definition{
		Type:       "object",
		Properties: make(map[string]Property),
	}

	for _, v2 := range si.Items {
		if v2.TypeRef != nil {
			if v2.IsArray {
				p := Property{}
				if v2.IsTDArray {
					p = Property{
						Type:        "array",
						Description: v2.Note,
						Items: &PropertyItems{
							Type: "array",
							Items: map[string]string{
								"type": GetKvType(v2.Type, false, true),
							},
						},
					}
				} else {
					p = Property{
						Type:        "array",
						Description: v2.Note,
						Items: &PropertyItems{
							Type:   GetKvType(v2.Type, v2.IsArray, true),
							Format: GetKvType(v2.Type, v2.IsArray, false),
							Ref:    doc.SetDefinition(m, v2.TypeRef),
						},
					}
				}
				def.Properties[v2.Name] = p

			} else {
				def.Properties[v2.Name] = Property{
					Type:   GetKvType(v2.Type, v2.IsArray, true),
					Format: GetKvType(v2.Type, v2.IsArray, false),
					Ref:    doc.SetDefinition(m, v2.TypeRef),
					Items:  nil,
				}
			}

		} else {
			if v2.IsArray {
				p := Property{}
				if v2.IsTDArray {
					p = Property{
						Type:        "array",
						Description: v2.Note,
						Items: &PropertyItems{
							Type: "array",
							Items: map[string]string{
								"type":   GetKvType(v2.Type, false, true),
								"format": v2.Type,
							},
						},
					}
				} else {
					p = Property{
						Type:        "array",
						Description: v2.Note,
						Items: &PropertyItems{
							Type:   GetKvType(v2.Type, v2.IsArray, true),
							Format: GetKvType(v2.Type, v2.IsArray, false),
							Ref:    doc.SetDefinition(m, v2.TypeRef),
						},
					}
				}

				def.Properties[v2.Name] = p
			} else {
				if v2.IsQuery || v2.IsHeader || v2.IsPath || v2.IsFormData {

				} else {
					def.Properties[v2.Name] = Property{
						Type:        GetKvType(v2.Type, v2.IsArray, true),
						Format:      GetKvType(v2.Type, v2.IsArray, false),
						Description: v2.Note,
						Items:       nil,
					}

				}
			}

		}
	}
	doc.AddDefinitions(si.Name, def)
	return "#/definitions/" + si.Name
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
		return "array"
	}

	if isType {
		if kt, ok := kvType[k]; ok {
			return kt
		}
		return "object"
	}
	if kf, ok := kvFormat[k]; ok {
		return kf
	}
	return k
}
