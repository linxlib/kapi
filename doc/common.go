package doc

import (
	"gitee.com/kirile/kapi/doc/swagger"
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

			// required
			tagStr = tag.Get("binding")
			tagStrs = strings.Split(tagStr, ",")
			for _, v := range tagStrs {
				if strings.EqualFold(v, "required") {
					info.Items[i].Required = true
					break
				}
			}
			// ---------------end

			// default
			info.Items[i].Default = tag.Get("default")
			// ---------------end
			//query
			v, b := tag.Lookup("query")
			//info.Items[i].IsQuery = b
			if b {
				info.Items[i].ParamType = ParamTypeQuery
				info.Items[i].Name = v
			}

			//header
			v, b = tag.Lookup("header")
			info.Items[i].IsHeader = b
			if b {
				info.Items[i].ParamType = ParamTypeHeader
				info.Items[i].Name = v
				info.Items[i].Required = true
			}
			//formData
			v, b = tag.Lookup("form")

			info.Items[i].IsFormData = b
			if b {
				info.Items[i].ParamType = ParamTypeForm
				info.Items[i].IsHeader = false

				info.Items[i].Name = v
			}
			//path
			v, b = tag.Lookup("path")
			info.Items[i].IsPath = b
			if b {
				info.Items[i].ParamType = ParamTypePath
				info.Items[i].IsHeader = false
				info.Items[i].IsFormData = false
				//info.Items[i].Required = true
				info.Items[i].Name = v
			} else {
				v, b = tag.Lookup("uri")
				info.Items[i].IsPath = b
				if b {
					info.Items[i].IsHeader = false
					info.Items[i].IsFormData = false
					//info.Items[i].Required = true
					info.Items[i].Name = v
				}
			}

			////required
			//v,b = tag.Lookup("required")
			//if b {
			//	info.Items[i].Required = conv.Bool(v)
			//}

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
					p := swagger.Property{
						Type:        "array",
						Description: v2.Note,
						Items: &swagger.PropertyItems{
							Type:   swagger.GetKvType(v2.Type, v2.IsArray, true),
							Format: swagger.GetKvType(v2.Type, v2.IsArray, false),
							Ref:    m.SetDefinition(doc, v2.TypeRef),
						},
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
		doc.AddDefinitions(tmp.Name, def)
		return "#/definitions/" + tmp.Name
	}
	return ""
}
