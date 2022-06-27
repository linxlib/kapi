package kapi

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

// ApiFunc Custom context support
type ApiFunc func(*gin.Context) interface{}

// RecoverErrorFunc recover 错误设置
type RecoverErrorFunc func(interface{})

// store the comment for the controller method. 生成注解路由
type genComment struct {
	RouterPath   string
	IsDeprecated bool
	ResultType   string
	Summary      string //方法说明
	Description  string // 方法注释
	Methods      []string
	TokenHeader  string
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
)
