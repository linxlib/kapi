package swagger

import (
	"encoding/json"
	"github.com/linxlib/kapi/internal"
	doc2 "github.com/linxlib/kapi/internal/doc"
	"strings"
)

// DocSwagger ...
type DocSwagger struct {
	Client *APIBody
}

func New(docName string, docVer string, docDesc string) *DocSwagger {
	ds := &DocSwagger{}
	ds.Client = &APIBody{
		Head: Head{Swagger: version},
		Info: Info{
			Description: docDesc,
			Version:     docVer,
			Title:       docName,
		},
		Host:         "",
		BasePath:     "",
		Schemes:      []string{},
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
	b, _ := json.MarshalIndent(doc.Client, "", "     ")
	return string(b)
}

func (doc *DocSwagger) SetDefinition(m *doc2.Model, si *doc2.StructInfo) string {
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
								"type": internal.GetKvType(v2.Type, false, true),
							},
						},
					}
				} else {
					p = Property{
						Type:        "array",
						Description: v2.Note,
						Items: &PropertyItems{
							Type:   internal.GetKvType(v2.Type, v2.IsArray, true),
							Format: internal.GetKvType(v2.Type, v2.IsArray, false),
							Ref:    doc.SetDefinition(m, v2.TypeRef),
						},
					}
				}
				def.Properties[v2.Name] = p

			} else {
				def.Properties[v2.Name] = Property{
					Type:   internal.GetKvType(v2.Type, v2.IsArray, true),
					Format: internal.GetKvType(v2.Type, v2.IsArray, false),
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
								"type":   internal.GetKvType(v2.Type, false, true),
								"format": v2.Type,
							},
						},
					}
				} else {
					p = Property{
						Type:        "array",
						Description: v2.Note,
						Items: &PropertyItems{
							Type:   internal.GetKvType(v2.Type, v2.IsArray, true),
							Format: internal.GetKvType(v2.Type, v2.IsArray, false),
							Ref:    doc.SetDefinition(m, v2.TypeRef),
						},
					}
				}

				def.Properties[v2.Name] = p
			} else {
				if !v2.IsQuery && !v2.IsHeader && !v2.IsPath && !v2.IsFormData {
					def.Properties[v2.Name] = Property{
						Type:        internal.GetKvType(v2.Type, v2.IsArray, true),
						Format:      internal.GetKvType(v2.Type, v2.IsArray, false),
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
