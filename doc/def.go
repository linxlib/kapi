package doc

// ElementInfo 结构信息
type ElementInfo struct {
	Name      string      // 参数名
	Tag       string      // 标签
	Type      string      // 类型
	TypeRef   *StructInfo // 类型定义
	IsArray   bool        // 是否是数组
	Required  bool        // 是否必须
	Note      string      // 注释
	Default   string      // 默认值
	ParamType ParamType

	IsQuery    bool // 是否是query
	IsHeader   bool // 是否是header
	IsFormData bool // 是否是表单参数
	IsPath     bool // 是否是路径参数
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
	Items []ElementInfo // 结构体元素
	Note  string        // 注释
	Name  string        // 结构体名字
	Pkg   string        // 包名
}

// DocModel Model
type DocModel struct {
	RouterPath string
	Methods    []string
	Note       string
	Req, Resp  *StructInfo
}
