package kapi

type IResultBuilder interface {
	OnSuccess(msg string, data any) (statusCode int, result any)
	OnData(msg string, count int64, data any) (statusCode int, result any)
	OnFail(msg string, data any) (statusCode int, result any)
	OnError(msg string, err error) (statusCode int, result any)
	OnErrorDetail(msg string, err any) (statusCode int, result any)
	OnUnAuthed(msg string) (statusCode int, result any)
	OnNoPermission(msg string) (statusCode int, result any)
	OnNotFound(msg string) (statusCode int, result any)
}

var _ IResultBuilder = (*ResultBuilderUnimplemented)(nil)

type ResultBuilderUnimplemented struct{}

func (r ResultBuilderUnimplemented) OnData(msg string, count int64, data any) (statusCode int, result any) {
	return 0, nil
}

func (r ResultBuilderUnimplemented) OnErrorDetail(msg string, err any) (statusCode int, result any) {
	return 0, nil
}

func (r ResultBuilderUnimplemented) OnSuccess(msg string, data any) (statusCode int, result any) {
	return 0, nil
}

func (r ResultBuilderUnimplemented) OnFail(msg string, data any) (statusCode int, result any) {
	return 0, nil
}

func (r ResultBuilderUnimplemented) OnError(msg string, err error) (statusCode int, result any) {
	return 0, nil
}

func (r ResultBuilderUnimplemented) OnUnAuthed(msg string) (statusCode int, result any) {
	return 0, nil
}

func (r ResultBuilderUnimplemented) OnNoPermission(msg string) (statusCode int, result any) {
	return 0, nil
}

func (r ResultBuilderUnimplemented) OnNotFound(msg string) (statusCode int, result any) {
	return 0, nil
}

// RegisterResultBuilder 注册返回body的builder处理类
//
//	@param builder 实现 IResultBuilder 接口
func (b *KApi) RegisterResultBuilder(builder IResultBuilder) {
	b.MapTo(builder, (*IResultBuilder)(nil))
}

var _ IResultBuilder = (*DefaultResultBuilder)(nil)

type DefaultResultBuilder struct {
}

func (d DefaultResultBuilder) OnData(msg string, count int64, data any) (statusCode int, result any) {
	return 200, messageBody{
		Code:  "SUCCESS",
		Msg:   msg,
		Count: count,
		Data:  data,
	}
}

func (d DefaultResultBuilder) OnErrorDetail(msg string, err any) (statusCode int, result any) {
	return 500, messageBody{
		Code: "FAIL",
		Msg:  msg,
		Data: err,
	}
}

func (d DefaultResultBuilder) OnSuccess(msg string, data any) (statusCode int, result any) {
	return 200, messageBody{
		Code: "SUCCESS",
		Msg:  msg,
		Data: data,
	}
}

func (d DefaultResultBuilder) OnFail(msg string, data any) (statusCode int, result any) {
	return 400, messageBody{
		Code: "FAIL",
		Msg:  msg,
		Data: data,
	}
}

func (d DefaultResultBuilder) OnError(msg string, err error) (statusCode int, result any) {
	return 500, messageBody{
		Code: "FAIL",
		Msg:  msg,
		Data: err.Error(),
	}
}

func (d DefaultResultBuilder) OnUnAuthed(msg string) (statusCode int, result any) {
	return 401, messageBody{
		Code: "FAIL",
		Msg:  msg,
	}
}

func (d DefaultResultBuilder) OnNoPermission(msg string) (statusCode int, result any) {
	return 403, messageBody{
		Code: "FAIL",
		Msg:  msg,
	}
}

func (d DefaultResultBuilder) OnNotFound(msg string) (statusCode int, result any) {
	return 404, messageBody{
		Code: "FAIL",
		Msg:  msg,
	}
}

// messageBody 自定义的响应body类型
type messageBody struct {
	Code  interface{} `json:"code"`
	Msg   string      `json:"msg"`
	Count int64       `json:"count,omitempty"`
	Data  interface{} `json:"data"`
}

// WriteJSON 写入json对象
func (c *Context) WriteJSON(obj interface{}) {
	c.PureJSON(200, obj)
}

func (c *Context) writeMessage(msg string) {
	c.PureJSON(c.ResultBuilder.OnSuccess(msg, nil))
}
func (c *Context) writeError(err error) {
	c.PureJSON(c.ResultBuilder.OnError(err.Error(), err))
}
func (c *Context) writeErrorDetail(err interface{}) {
	c.PureJSON(c.ResultBuilder.OnErrorDetail("", err))
}
func (c *Context) writeFailMsg(msg string) {
	c.PureJSON(c.ResultBuilder.OnFail(msg, nil))
}

func (c *Context) writeList(count int64, list interface{}) {
	c.PureJSON(c.ResultBuilder.OnData("", count, list))
}
func (c *Context) writeNoPermissionMsg(msg string) {
	c.PureJSON(c.ResultBuilder.OnNoPermission(msg))
}
func (c *Context) writeNotFoundMsg(msg string) {
	c.PureJSON(c.ResultBuilder.OnNotFound(msg))
}
func (c *Context) writeUnAuthedMsg(msg string) {
	c.PureJSON(c.ResultBuilder.OnUnAuthed(msg))
}
func (c *Context) writeMsgAndData(msg string, data interface{}) {
	c.PureJSON(c.ResultBuilder.OnData(msg, 0, data))
}

var KAPIEXIT = "kapiexit"

func (c *Context) Exit() {
	panic(KAPIEXIT)
}

func (c *Context) ListExit(count int64, list interface{}) {
	c.writeList(count, list)
	c.Exit()
}

func (c *Context) DataExit(data interface{}) {
	c.writeList(0, data)
	c.Exit()
}
func (c *Context) MessageAndDataExit(message string, data interface{}) {
	c.writeMsgAndData(message, data)
	c.Exit()
}

func (c *Context) Success(data ...interface{}) {
	if len(data) > 0 {
		switch data[0].(type) {
		case string:
			c.writeMessage(data[0].(string))
		default:
			c.writeList(0, data[0])
		}
	} else {
		c.writeMessage("")
	}
}

func (c *Context) SuccessExit(data ...interface{}) {
	if len(data) > 0 {
		switch data[0].(type) {
		case string:
			c.writeMessage(data[0].(string))
			c.Exit()
		default:
			c.writeList(0, data[0])
			c.Exit()
		}
	} else {
		c.writeMessage("")
		c.Exit()
	}
}
func (c *Context) NoPermissionExit(msg ...string) {
	if len(msg) > 0 {
		c.writeNoPermissionMsg(msg[0])
		c.Exit()
	} else {
		c.writeNoPermissionMsg("no permission")
		c.Exit()
	}
}

func (c *Context) NotFoundExit(msg ...string) {
	if len(msg) > 0 {
		c.writeNotFoundMsg(msg[0])
		c.Exit()
	} else {
		c.writeNotFoundMsg("not found")
		c.Exit()
	}
}

func (c *Context) UnAuthed(msg ...string) {
	if len(msg) > 0 {
		c.writeUnAuthedMsg(msg[0])
	} else {
		c.writeUnAuthedMsg("un authed")
	}
	c.Exit()
}

func (c *Context) NoPerm(msg ...string) {
	if len(msg) > 0 {
		c.writeNoPermissionMsg(msg[0])
	} else {
		c.writeNoPermissionMsg("no permission")
	}
	c.Exit()
}

func (c *Context) FailAndExit(data ...interface{}) {
	if len(data) > 0 {
		switch data[0].(type) {
		case string:
			c.writeFailMsg(data[0].(string))
			c.Exit()
		case error:
			c.writeError(data[0].(error))
			c.Exit()
		default:
			c.writeErrorDetail(data[0])
			c.Exit()
		}
	} else {
		c.writeFailMsg("action failed")
		c.Exit()
	}
}
