package kapi

import (
	"github.com/gin-gonic/gin"
	"github.com/linxlib/conv"
	"github.com/linxlib/kapi/inject"
	"strings"
)

// Context Wrapping gin context to custom context
type Context struct { // 包装gin的上下文到自定义context
	*gin.Context
	inj inject.Injector
}

// newContext Create a new custom context
func newContext(c *gin.Context) *Context { // 新建一个自定义context
	return &Context{
		inj:     inject.New(),
		Context: c,
	}
}

func (c *Context) Method() string {
	return c.Request.Method
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

// RemoteAddr returns more real IP address.
func (c *Context) RemoteAddr() string {
	addr := c.Request.Header.Get("X-Real-IP")
	if len(addr) == 0 {
		addr = c.Request.Header.Get("X-Forwarded-For")
		if addr == "" {
			addr = c.Request.RemoteAddr
			if i := strings.LastIndex(addr, ":"); i > -1 {
				addr = addr[:i]
			}
		}
	}
	return addr
}

func (c *Context) Map(i ...interface{}) inject.TypeMapper {
	return c.inj.Map(i...)
}

func (c *Context) MapTo(i interface{}, j interface{}) inject.TypeMapper {
	return c.inj.MapTo(i, j)
}
