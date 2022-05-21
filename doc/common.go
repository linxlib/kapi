package doc

import (
	"github.com/linxlib/kapi/doc/swagger"
)

func (m *Model) analysisStructInfo(info *StructInfo) {
	if info != nil {
		for _, item := range info.Items {
			item.execute()
			if item.TypeRef != nil {
				m.analysisStructInfo(item.TypeRef)
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
