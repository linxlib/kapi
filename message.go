package kapi

// messageBody 自定义的响应body类型
type messageBody struct {
	Code  interface{} `json:"code"`
	Msg   string      `json:"msg"`
	Count int64       `json:"count,omitempty"`
	Data  interface{} `json:"data"`
}

type RESULT_CODE int

const (
	RESULT_CODE_SUCCESS RESULT_CODE = iota //SuccessExit DataExit ListExit 时
	RESULT_CODE_FAIL                       // FailExit
	RESULT_CODE_ERROR                      //FailExit时传参error
	RESULT_CODE_UNAUTHED
)

// RegisterFuncGetResult 注册返回json结构的func
func RegisterFuncGetResult(i FuncGetResult) {
	DefaultGetResult = i
}

type FuncGetResult = func(code RESULT_CODE, msg string, count int64, data interface{}) (int, interface{})

var DefaultGetResult = _defaultGetResult

func _defaultGetResult(code RESULT_CODE, msg string, count int64, data interface{}) (int, interface{}) {
	if code == RESULT_CODE_SUCCESS {
		return 200, messageBody{
			Code:  "SUCCESS",
			Msg:   msg,
			Count: count,
			Data:  data,
		}
	} else if code == RESULT_CODE_ERROR {
		return 500, messageBody{
			Code:  "FAIL",
			Msg:   msg,
			Count: count,
			Data:  data,
		}
	} else if code==RESULT_CODE_UNAUTHED{
		return 401, messageBody{
			Code:  "FAIL",
			Msg:   msg,
			Count: count,
			Data:  data,
		}
	} else {
		return 400, messageBody{
			Code:  "FAIL",
			Msg:   msg,
			Count: count,
			Data:  data,
		}
	}
}

//WriteJSON 写入json对象
func (c *Context) WriteJSON(obj interface{}) {
	c.JSON(200, obj)
}

func (c *Context) writeMessage(msg string) {
	c.JSON(DefaultGetResult(RESULT_CODE_SUCCESS, msg, 0, nil))
}
func (c *Context) writeError(err error) {
	c.JSON(DefaultGetResult(RESULT_CODE_ERROR, err.Error(), 0, nil))
}
func (c *Context) writeErrorDetail(err interface{}) {
	c.JSON(DefaultGetResult(RESULT_CODE_ERROR, "", 0, err))
}
func (c *Context) writeFailMsg(msg string) {
	c.JSON(DefaultGetResult(RESULT_CODE_FAIL, msg, 0, nil))
}
func (c *Context) writeList(count int64, list interface{}) {
	c.JSON(DefaultGetResult(RESULT_CODE_SUCCESS, "", count, list))
}

func (c *Context) ListExit(count int64, list interface{}) {
	c.writeList(count, list)
	c.Exit()
}

func (c *Context) DataExit(data interface{}) {
	c.writeList(0, data)
	c.Exit()
}

func (c *Context) SuccessExit(data ...interface{}) {
	if len(data) > 0 {
		switch data[0].(type) {
		case string:
			{
				c.writeMessage(data[0].(string))
				c.Exit()
			}
		default:
			c.writeList(0, data[0])
			c.Exit()
		}
	} else {
		c.writeMessage("")
		c.Exit()
	}
}

func (c *Context) FailAndExit(data ...interface{}) {
	if len(data) > 0 {
		switch data[0].(type) {
		case string:
			{
				c.writeFailMsg(data[0].(string))
				c.Exit()
			}
		case error:
			{
				c.writeError(data[0].(error))
				c.Exit()
			}

		default:
			c.writeErrorDetail(data[0])
			c.Exit()
		}
	} else {
		c.writeFailMsg("操作失败")
		c.Exit()
	}
}


