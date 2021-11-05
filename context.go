package kapi

import (
	"github.com/gin-gonic/gin"
	"github.com/linxlib/conv"
)

// Context Wrapping gin context to custom context
type Context struct { // 包装gin的上下文到自定义context
	*gin.Context
}

// GetVersion Get the version by req url
func (c *Context) GetVersion() string { // 获取版本号
	return c.Param("version")
}

// NewCtx Create a new custom context
func NewCtx(c *gin.Context) *Context { // 新建一个自定义context
	return &Context{c}
}

// NewAPIFunc default of custom handlefunc
func NewAPIFunc(c *gin.Context) interface{} {
	return NewCtx(c)
}

var KAPIEXIT = "kapiexit"

func (c *Context) Exit() {
	panic(KAPIEXIT)
}

func (c *Context) GetQueryString(key string) string {
	tmp, b := c.GetQuery(key)
	if b {
		return tmp
	} else {
		return ""
	}
}

func (c *Context) GetQueryInt(key string, def ...int) int {
	tmp, b := c.GetQuery(key)
	if b {
		return conv.Int(tmp)
	} else {
		if len(def) > 0 {
			return def[0]
		}
		return -1
	}
}

func (c *Context) GetQueryInt64(key string) int64 {
	tmp, b := c.GetQuery(key)
	if b {
		return conv.Int64(tmp)
	} else {
		return -1
	}
}

func (c *Context) GetQueryUInt64(key string) uint64 {
	tmp, b := c.GetQuery(key)
	if b {
		return conv.Uint64(tmp)
	} else {
		return 0
	}
}

func (c *Context) GetQueryUInt(key string) uint {
	tmp, b := c.GetQuery(key)
	if b {
		return conv.Uint(tmp)
	} else {
		return 0
	}
}

func (c *Context) GetQueryBool(key string) bool {
	tmp, b := c.GetQuery(key)
	if b {
		return conv.Bool(tmp)
	} else {
		return false
	}
}

func (c *Context) GetPageSize() (int, int) {
	page := c.GetQueryInt("page", 1)
	size := c.GetQueryInt("size", 10)
	return conv.Int(page) - 1, conv.Int(size)
}

func (c *Context) SaveFile(form string, dst string) error {
	f, err := c.FormFile(form)
	if err != nil {
		return err
	}
	err = c.SaveUploadedFile(f, dst)
	if err != nil {
		return err
	}
	return nil
}
