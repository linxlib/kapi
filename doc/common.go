package doc

import (
	"github.com/linxlib/kapi/doc/swagger"
	"reflect"
	"strings"
)

func (m *Model) analysisStructInfo(info *StructInfo) {
	if info != nil {
		for i := 0; i < len(info.Items); i++ {
			tag := reflect.StructTag(strings.Trim(info.Items[i].Tag, "`"))
			info.Items[i].IsQuery = false
			info.Items[i].IsHeader = false
			info.Items[i].IsFormData = false
			info.Items[i].IsPath = false
			// json
			tagStr := tag.Get("json")
			if tagStr == "-" || tagStr == "" {
				tagStr = tag.Get("url")
			}
			tagStrs := strings.Split(tagStr, ",")
			if len(tagStrs[0]) > 0 {
				info.Items[i].ParamType = -1
				info.Items[i].Name = tagStrs[0]
			}
			// -------- end

			// 必填
			tagStr = tag.Get("binding")
			tagStrs = strings.Split(tagStr, ",")
			for _, v := range tagStrs {
				if strings.EqualFold(v, "required") {
					info.Items[i].Required = true
					break
				}
			}
			// 默认值
			info.Items[i].Default = tag.Get("default")
			//query
			v, b := tag.Lookup("query")
			if b {
				info.Items[i].ParamType = ParamTypeQuery
				info.Items[i].Name = v
			}
			//请求头header
			v, b = tag.Lookup("header")
			info.Items[i].IsHeader = b
			if b {
				info.Items[i].ParamType = ParamTypeHeader
				info.Items[i].Name = v
				info.Items[i].Required = true
			}
			//表单 formData
			v, b = tag.Lookup("form")

			info.Items[i].IsFormData = b
			if b {
				info.Items[i].ParamType = ParamTypeForm
				info.Items[i].IsHeader = false
				info.Items[i].Name = v
			}
			//url path
			v, b = tag.Lookup("path")
			info.Items[i].IsPath = b
			if b {
				info.Items[i].ParamType = ParamTypePath
				info.Items[i].IsHeader = false
				info.Items[i].IsFormData = false
				info.Items[i].Name = v
			} else {
				v, b = tag.Lookup("uri")
				info.Items[i].IsPath = b
				if b {
					info.Items[i].IsHeader = false
					info.Items[i].IsFormData = false
					info.Items[i].Name = v
				}
			}

			if info.Items[i].TypeRef != nil {
				m.analysisStructInfo(info.Items[i].TypeRef)
			}
		}

	}
}

func (m *Model) SetDefinition(doc *swagger.DocSwagger, tmp *StructInfo) string {
	if tmp != nil {
		var def swagger.Definition
		def.Type = "object"
		def.Properties = make(map[string]swagger.Property)
		for _, v2 := range tmp.Items {
			if v2.TypeRef != nil {
				if v2.IsArray {
					p := swagger.Property{}
					if v2.IsTDArray {
						p = swagger.Property{
							Type:        "array",
							Description: v2.Note,
							Items: &swagger.PropertyItems{
								Type: "array",
								Items: map[string]string{
									"type": swagger.GetKvType(v2.Type, false, true),
								},
							},
						}
					} else {
						p = swagger.Property{
							Type:        "array",
							Description: v2.Note,
							Items: &swagger.PropertyItems{
								Type:   swagger.GetKvType(v2.Type, v2.IsArray, true),
								Format: swagger.GetKvType(v2.Type, v2.IsArray, false),
								Ref:    m.SetDefinition(doc, v2.TypeRef),
							},
						}
					}

					def.Properties[v2.Name] = p
				} else {
					def.Properties[v2.Name] = swagger.Property{
						Type:   swagger.GetKvType(v2.Type, v2.IsArray, true),
						Format: swagger.GetKvType(v2.Type, v2.IsArray, false),
						Ref:    m.SetDefinition(doc, v2.TypeRef),
						Items:  nil,
					}
				}

			} else {
				if v2.IsArray {
					p := swagger.Property{}
					if v2.IsTDArray {
						p = swagger.Property{
							Type:        "array",
							Description: v2.Note,
							Items: &swagger.PropertyItems{
								Type: "array",
								Items: map[string]string{
									"type":   swagger.GetKvType(v2.Type, false, true),
									"format": v2.Type,
								},
							},
						}
					} else {
						p = swagger.Property{
							Type:        "array",
							Description: v2.Note,
							Items: &swagger.PropertyItems{
								Type:   swagger.GetKvType(v2.Type, v2.IsArray, true),
								Format: swagger.GetKvType(v2.Type, v2.IsArray, false),
								Ref:    m.SetDefinition(doc, v2.TypeRef),
							},
						}
					}

					def.Properties[v2.Name] = p
				} else {
					if v2.IsQuery || v2.IsHeader || v2.IsPath || v2.IsFormData {

					} else {
						def.Properties[v2.Name] = swagger.Property{
							Type:        swagger.GetKvType(v2.Type, v2.IsArray, true),
							Format:      swagger.GetKvType(v2.Type, v2.IsArray, false),
							Description: v2.Note,
							Items:       nil,
						}
					}
				}

			}
		}
		doc.AddDefinitions(tmp.Name, def)
		return "#/definitions/" + tmp.Name
	}
	return ""
}
