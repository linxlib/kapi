package kapi

// IOnSuccess 200
type IOnSuccess = func(msg string, data any) (statusCode int, result any)

// IOnFail 400
type IOnFail = func(msg string, data any) (statusCode int, result any)

// IOnData 200
type IOnData = func(msg string, count int64, data any) (statusCode int, result any)

// IOnError 500
type IOnError = func(msg string, err error) (statusCode int, result any)
type IOnErrorDetail = func(msg string, err any) (statusCode int, result any)

// IOnUnAuthed 401
type IOnUnAuthed = func(msg string) (statusCode int, result any)

// IOnNoPermission 403
type IOnNoPermission = func(msg string) (statusCode int, result any)

// IOnNotFound 404
type IOnNotFound = func(msg string) (statusCode int, result any)

func (b *KApi) RewriteOnSuccess(builder IOnSuccess) {
	b.MapTo(builder, (*IOnSuccess)(nil))
}
func (b *KApi) RewriteOnFail(builder IOnFail) {
	b.MapTo(builder, (*IOnFail)(nil))
}
func (b *KApi) RewriteOnErrorDetail(builder IOnErrorDetail) {
	b.MapTo(builder, (*IOnErrorDetail)(nil))
}
func (b *KApi) RewriteOnError(builder IOnError) {
	b.MapTo(builder, (*IOnError)(nil))
}
func (b *KApi) RewriteOnUnAuthed(builder IOnUnAuthed) {
	b.MapTo(builder, (*IOnUnAuthed)(nil))
}
func (b *KApi) RewriteOnData(builder IOnData) {
	b.MapTo(builder, (*IOnData)(nil))
}
func (b *KApi) RewriteOnNoPermission(builder IOnNoPermission) {
	b.MapTo(builder, (*IOnNoPermission)(nil))
}
func (b *KApi) RewriteOnNotFound(builder IOnNotFound) {
	b.MapTo(builder, (*IOnNotFound)(nil))
}

type DefaultResultBuilder struct {
	OnSuccess      IOnSuccess
	OnFail         IOnFail
	OnNotFound     IOnNotFound
	OnNoPermission IOnNoPermission
	OnData         IOnData
	OnError        IOnError
	OnUnAuthed     IOnUnAuthed
	OnErrorDetail  IOnErrorDetail
}

func NewDefaultBuilder() *DefaultResultBuilder {
	return &DefaultResultBuilder{
		OnSuccess: func(msg string, data any) (statusCode int, result any) {
			return 200, messageBody{
				Code: -1,
				Msg:  msg,
				Data: data,
			}
		},
		OnFail: func(msg string, data any) (statusCode int, result any) {
			return 400, messageBody{
				Code: -1,
				Msg:  msg,
				Data: data,
			}
		},
		OnData: func(msg string, count int64, data any) (statusCode int, result any) {
			return 200, messageBody{
				Code:  0,
				Msg:   msg,
				Count: count,
				Data:  data,
			}
		},
		OnError: func(msg string, err error) (statusCode int, result any) {
			return 500, messageBody{
				Code: -1,
				Msg:  msg,
				Data: err.Error(),
			}
		},
		OnNoPermission: func(msg string) (statusCode int, result any) {
			return 403, messageBody{
				Code: -1,
				Msg:  msg,
			}
		},
		OnNotFound: func(msg string) (statusCode int, result any) {
			return 404, messageBody{
				Code: -1,
				Msg:  msg,
			}
		},
		OnUnAuthed: func(msg string) (statusCode int, result any) {
			return 401, messageBody{
				Code: -1,
				Msg:  msg,
			}
		},
		OnErrorDetail: func(msg string, err any) (statusCode int, result any) {
			return 500, messageBody{
				Code: -1,
				Msg:  msg,
				Data: err,
			}
		},
	}
}

// messageBody 自定义的响应body类型
type messageBody struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Count int64       `json:"count,omitempty"`
	Data  interface{} `json:"data"`
}

// WriteJSON 写入json对象
func (c *Context) WriteJSON(obj interface{}) {
	c.PureJSON(200, obj)
}

func (c *Context) writeMessage(msg string) {
	c.PureJSON(c.OnSuccess(msg, nil))
}
func (c *Context) writeError(err error) {
	c.PureJSON(c.OnError(err.Error(), err))
}
func (c *Context) writeErrorDetail(err interface{}) {
	c.PureJSON(c.OnErrorDetail("", err))
}
func (c *Context) writeFailMsg(msg string) {
	c.PureJSON(c.OnFail(msg, nil))
}

func (c *Context) writeList(count int64, list any) {
	c.PureJSON(c.OnData("", count, list))
}
func (c *Context) writeNoPermissionMsg(msg string) {
	c.PureJSON(c.OnNoPermission(msg))
}
func (c *Context) writeNotFoundMsg(msg string) {
	c.PureJSON(c.OnNotFound(msg))
}
func (c *Context) writeUnAuthedMsg(msg string) {
	c.PureJSON(c.OnUnAuthed(msg))
}
func (c *Context) writeMsgAndData(msg string, data any) {
	c.PureJSON(c.OnData(msg, 0, data))
}

var KAPIEXIT = "kapiexit"

func (c *Context) Exit() {
	panic(KAPIEXIT)
}

func (c *Context) ListExit(count int64, list any) {
	c.writeList(count, list)
	c.Exit()
}

func (c *Context) DataExit(data any) {
	c.writeList(0, data)
	c.Exit()
}
func (c *Context) MessageAndDataExit(message string, data any) {
	c.writeMsgAndData(message, data)
	c.Exit()
}

func (c *Context) Success(data ...any) {
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

func (c *Context) SuccessExit(data ...any) {
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

func (c *Context) FailAndExit(data ...any) {
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
