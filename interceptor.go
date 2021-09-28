package kapi

import (
	"context"
	"github.com/gin-gonic/gin"
	"time"
)

// InterceptorContext 对象调用前后执行中间件参数
type InterceptorContext struct {
	C        *gin.Context
	FuncName string          // 函数名
	Req      interface{}     // 调用前的请求参数
	Resp     interface{}     // 调用后的响应参数
	Error    error           // 错误信息
	Context  context.Context // 占位上下文参数，可用于存储其他参数，前后连接可用
}

// Interceptor 对象调用前后执行中间件(支持总的跟对象单独添加)
type Interceptor interface {
	GinBefore(req *InterceptorContext) bool
	GinAfter(req *InterceptorContext) bool
}

// DefaultGinBeforeAfter 默认 BeforeAfter Middleware
type DefaultGinBeforeAfter struct {
}

type timeTrace struct{}

// GinBefore call之前调用
func (d *DefaultGinBeforeAfter) GinBefore(req *InterceptorContext) bool {
	req.Context = context.WithValue(req.Context, timeTrace{}, time.Now())
	return true
}

// GinAfter call之后调用
func (d *DefaultGinBeforeAfter) GinAfter(req *InterceptorContext) bool {
	begin := (req.Context.Value(timeTrace{})).(time.Time)
	now := time.Now()
	_log.Infof("[middleware] call[%v] [%v]", req.FuncName, now.Sub(begin))
	if req.Error != nil {
		_, req.Resp = defaultGetResult(RESULT_CODE_ERROR, req.Error.Error(), 0, nil)
	} else {
		_, req.Resp = defaultGetResult(RESULT_CODE_SUCCESS, "", 0, req.Resp)
	}
	return true
}
