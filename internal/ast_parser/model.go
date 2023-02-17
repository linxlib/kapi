package ast_parser

import (
	"reflect"
)

type importItem struct {
	name       string //alias of import path
	importPath string
}

// File 文件
type File struct {
	Name    string        //文件名
	Imports []*importItem //导入
	PkgPath string        //包路径
	Structs []*Struct     //结构体
	Docs    []string      //注释
}

// Struct 结构体
type Struct struct {
	Name    string    //结构名称
	PkgPath string    //包路径
	Fields  []*Field  //字段
	Methods []*Method //方法
	Docs    []string  //上方文档注释
}

func (t *Struct) GetAllFieldsByTag(tag string) []*Field {
	rtn := make([]*Field, 0)
	for _, field := range t.Fields {
		if !field.hasTag && tag == "json" { //if no tag presents, act as json tag
			rtn = append(rtn, field)
			continue
		}
		if _, ok := field.Tag.Lookup(tag); ok {
			rtn = append(rtn, field)
		}
	}
	return rtn
}

// Method 结构体方法
type Method struct {
	Receiver  *Receiver //接收器
	PkgPath   string    //包路径
	Name      string    //方法名称
	Private   bool
	Signature string //方法签名
	Docs      []string
	Params    []*Param //函数参数
	Results   []*Param //函数返回值
}

// Receiver 接收器
type Receiver struct {
	Name    string
	Pointer bool
	Type    string
}

type Param struct {
	Name        string
	PkgPath     string //包路径
	Pointer     bool
	Slice       bool
	Type        string
	typeString  string
	ignoreParse bool
	Struct      *Struct //字段
}

type Field struct {
	Name    string //字段名
	PkgPath string //包路径
	Type    string //类型
	hasTag  bool
	Tag     reflect.StructTag //标签
	Private bool              //私有
	Pointer bool              //指针
	Slice   bool              //slice
	Docs    []string          //上方文档注释
	Comment string            //末尾的注释
}

func (f *Field) GetTag(tag string) string {
	return f.Tag.Get(tag)
}
