package doc

// StructInfo struct define
type StructInfo struct {
	Items   []*ElementInfo // 结构体元素
	IsArray bool
	File    string
	Note    string // 注释
	Name    string // 结构体名字
	Pkg     string // 包名
}
