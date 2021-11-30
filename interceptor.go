package kapi

import (
	"context"
	"gitee.com/kirile/kapi/internal"
	"time"
)

// InterceptorContext 对象调用前后执行中间件参数
type InterceptorContext struct {
	C        *Context
	FuncName string          // 函数名
	Req      interface{}     // 调用前的请求参数
	Resp     interface{}     // 调用后的响应参数
	RespCode int             //状态码
	Error    error           // 错误信息
	Context  context.Context // 占位上下文参数，可用于存储其他参数，前后连接可用
}

// Interceptor 对象调用前后执行中间件(支持总的跟对象单独添加)
type Interceptor interface {
	Before(req *InterceptorContext) bool
	After(req *InterceptorContext) bool
}

// DefaultBeforeAfter 默认 BeforeAfter Middleware
type DefaultBeforeAfter struct {
}

type timeTrace struct{}

// Before call之前调用
func (d *DefaultBeforeAfter) Before(req *InterceptorContext) bool {
	req.Context = context.WithValue(req.Context, timeTrace{}, time.Now())
	return true
}

// After call之后调用
func (d *DefaultBeforeAfter) After(req *InterceptorContext) bool {
	begin := (req.Context.Value(timeTrace{})).(time.Time)
	now := time.Now()
	internal.Log.Infof("[middleware] call[%v] [%v]", req.FuncName, now.Sub(begin))
	if req.Error != nil {
		_, req.Resp = GetResultFunc(RESULT_CODE_ERROR, req.Error.Error(), 0, nil)
	} else {
		_, req.Resp = GetResultFunc(RESULT_CODE_SUCCESS, "", 0, req.Resp)
	}
	return true
}
