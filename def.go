package kapi

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

// ApiFunc Custom context support
type ApiFunc func(*gin.Context) interface{}

// RecoverErrorFunc recover 错误设置
type RecoverErrorFunc func(interface{})

// paramInfo 参数类型描述
type paramInfo struct {
	Pkg    string // 包名
	Type   string // 类型
	Import string // import 包
}

// store the comment for the controller method. 生成注解路由
type genComment struct {
	RouterPath string
	Note       string // 注释
	Methods    []string
	TokenHeader string
}

// router style list.路由规则列表
type genRouterInfo struct {
	GenComment  genComment
	HandFunName string
}

type genInfo struct {
	List []genRouterInfo
	Tm   int64 //genout time
}

var (
	// Precompute the reflection type for error. Can't use error directly
	// because Typeof takes an empty interface value. This is annoying.
	typeOfError = reflect.TypeOf((*error)(nil)).Elem()

	genTemp = `//Package routers code generate by KApi on {{.T}}. do not edit it.
	package {{.PkgName}}
	
	import (
		"gitee.com/kirile/kapi"
	)
	 
	func init() {
		kapi.SetVersion({{.Tm}})
		{{range .List}}kapi.AddGenOne("{{.HandFunName}}", "{{.GenComment.RouterPath}}", []string{ {{getStringList .GenComment.Methods}} })
		{{end}} 
	}
	`
)
