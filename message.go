package kapi

// MessageBody 自定义的响应body类型
type MessageBody struct {
	Code  interface{} `json:"code"`
	Msg   string      `json:"msg"`
	Count int64       `json:"count,omitempty"`
	Data  interface{} `json:"data"`
}

//WriteJSON 写入json对象
func (c *Context) WriteJSON(obj interface{}) {
	c.JSON(200, obj)
}

func (c *Context) writeMessage(msg string) {
	c.JSON(200, MessageBody{
		Code:  "SUCCESS",
		Msg:   msg,
		Count: 0,
		Data:  nil,
	})
}
func (c *Context) writeError(err error) {
	c.JSON(500, MessageBody{
		Code:  "FAIL",
		Msg:   err.Error(),
		Count: 0,
		Data:  nil,
	})
}
func (c *Context) writeErrorDetail(err interface{}) {
	c.JSON(200, MessageBody{
		Code:  "FAIL",
		Msg:   "",
		Count: 0,
		Data:  err,
	})
}
func (c *Context) writeFailMsg(msg string) {
	c.JSON(200, MessageBody{
		Code:  "FAIL",
		Msg:   msg,
		Count: 0,
		Data:  nil,
	})
}

func (c *Context) writeList(count int64, list interface{}) {
	c.JSON(200, MessageBody{
		Code:  "SUCCESS",
		Msg:   "",
		Count: count,
		Data:  list,
	})
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
