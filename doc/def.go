package doc

import (
	"reflect"
	"strings"
)

// ElementInfo 结构信息
type ElementInfo struct {
	Name      string      // 参数名
	Tag       string      // 标签
	Type      string      // 类型
	TypeRef   *StructInfo // 类型定义
	IsArray   bool        // 是否是数组
	IsTDArray bool        // 是否二维数组

	Required  bool   // 是否必须
	Note      string // 注释
	Default   string // 默认值
	ParamType ParamType

	IsQuery    bool // 是否是query
	IsHeader   bool // 是否是header
	IsFormData bool // 是否是表单参数
	IsPath     bool // 是否是路径参数
}

func (ei *ElementInfo) execute() {
	tag := reflect.StructTag(strings.Trim(ei.Tag, "`"))
	ei.IsQuery = false
	ei.IsHeader = false
	ei.IsFormData = false
	ei.IsPath = false
	// json
	tagStr := tag.Get("json")
	if tagStr == "-" || tagStr == "" {
		tagStr = tag.Get("url")
	}
	tagStrs := strings.Split(tagStr, ",")
	if len(tagStrs[0]) > 0 {
		ei.ParamType = -1
		ei.Name = tagStrs[0]
	}
	// -------- end

	// 必填
	tagStr = tag.Get("binding")
	tagStrs = strings.Split(tagStr, ",")
	for _, v := range tagStrs {
		if strings.EqualFold(v, "required") {
			ei.Required = true
			break
		}
	}
	// 默认值
	//ei.Default = tag.Get("default")
	//query
	v, b := tag.Lookup("query")
	if b {
		ei.ParamType = ParamTypeQuery
		tmp := strings.Split(v, ",")
		if len(tmp) > 1 {
			for _, s := range tmp {
				if strings.HasPrefix(s, "default=") {
					ei.Default = strings.TrimPrefix(s, "default=")
					break
				}
			}
			ei.Name = tmp[0]
		} else {
			ei.Name = tmp[0]
		}

	}
	//请求头header
	v, b = tag.Lookup("header")
	ei.IsHeader = b
	if b {
		ei.ParamType = ParamTypeHeader
		//ei.Name = v
		tmp := strings.Split(v, ",")
		if len(tmp) > 1 {
			for _, s := range tmp {
				if strings.HasPrefix(s, "default=") {
					ei.Default = strings.TrimPrefix(s, "default=")
					break
				}
			}
			ei.Name = tmp[0]
		} else {
			ei.Name = tmp[0]
		}
		ei.Required = true
	}
	//表单 formData
	v, b = tag.Lookup("form")

	ei.IsFormData = b
	if b {
		ei.ParamType = ParamTypeForm
		ei.IsHeader = false
		tmp := strings.Split(v, ",")
		if len(tmp) > 1 {
			for _, s := range tmp {
				if strings.HasPrefix(s, "default=") {
					ei.Default = strings.TrimPrefix(s, "default=")
					break
				}
			}
			ei.Name = tmp[0]
		} else {
			ei.Name = tmp[0]
		}
	}
	//url path
	v, b = tag.Lookup("path")
	ei.IsPath = b
	if b {
		ei.ParamType = ParamTypePath
		ei.IsHeader = false
		ei.IsFormData = false
		tmp := strings.Split(v, ",")
		if len(tmp) > 1 {
			for _, s := range tmp {
				if strings.HasPrefix(s, "default=") {
					ei.Default = strings.TrimPrefix(s, "default=")
					break
				}
			}
			ei.Name = tmp[0]
		} else {
			ei.Name = tmp[0]
		}
	} else {
		v, b = tag.Lookup("uri")
		ei.IsPath = b
		if b {
			ei.IsHeader = false
			ei.IsFormData = false
			tmp := strings.Split(v, ",")
			if len(tmp) > 1 {
				for _, s := range tmp {
					if strings.HasPrefix(s, "default=") {
						ei.Default = strings.TrimPrefix(s, "default=")
						break
					}
				}
				ei.Name = tmp[0]
			} else {
				ei.Name = tmp[0]
			}
		}
	}
}

type ParamType int

const (
	ParamTypeQuery ParamType = iota
	ParamTypeHeader
	ParamTypeForm
	ParamTypePath
)

// StructInfo struct define
type StructInfo struct {
	Items   []*ElementInfo // 结构体元素
	IsArray bool
	Note    string // 注释
	Name    string // 结构体名字
	Pkg     string // 包名
}

// DocModel Model
type DocModel struct {
	RouterPath   string
	Methods      []string
	Note         string
	Req, Resp    *StructInfo
	TokenHeader  string
	IsDeprecated bool
}
