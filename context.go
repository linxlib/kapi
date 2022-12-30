package kapi

import (
	"github.com/gin-gonic/gin"
	"github.com/linxlib/conv"
	"github.com/linxlib/kapi/inject"
	"strings"
)

// Context KApi Context
type Context struct {
	*gin.Context
	inj inject.Injector
}

// newContext create a new custom context
func newContext(c *gin.Context) *Context {
	return &Context{
		Context: c,
		inj:     inject.New(),
	}
}

// Method get HTTP method
//
//	@return string
func (c *Context) Method() string {
	return c.Request.Method
}

// GetQueryString get query param of string
//
//	@param key
//
//	@return string
func (c *Context) GetQueryString(key string) string {
	tmp, b := c.GetQuery(key)
	if b {
		return tmp
	} else {
		return ""
	}
}

// GetQueryInt get query param of integer
//
//	@param key
//	@param def default value
//
//	@return int
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

// GetQueryInt64 get query param of long
//
//	@param key
//	@param def default value
//
//	@return int64
func (c *Context) GetQueryInt64(key string) int64 {
	tmp, b := c.GetQuery(key)
	if b {
		return conv.Int64(tmp)
	} else {
		return -1
	}
}

// GetQueryUInt64 get query param of uint64
//
//	@param key
//	@param def default value
//
//	@return uint64
func (c *Context) GetQueryUInt64(key string) uint64 {
	tmp, b := c.GetQuery(key)
	if b {
		return conv.Uint64(tmp)
	} else {
		return 0
	}
}

// GetQueryUInt get query param of uint
//
//	@param key
//	@param def default value
//
//	@return uint
func (c *Context) GetQueryUInt(key string) uint {
	tmp, b := c.GetQuery(key)
	if b {
		return conv.Uint(tmp)
	} else {
		return 0
	}
}

// GetQueryBool get query param of boolean
//
//	@param key
//	@param def default value
//
//	@return bool
func (c *Context) GetQueryBool(key string) bool {
	tmp, b := c.GetQuery(key)
	if b {
		return conv.Bool(tmp)
	} else {
		return false
	}
}

// GetPageSize get query page and size
//
//	@return int page
//	@return int size
func (c *Context) GetPageSize() (int, int) {
	page := c.GetQueryInt("page", 1)
	size := c.GetQueryInt("size", 10)
	return conv.Int(page) - 1, conv.Int(size)
}

// SaveFile save form file to destination
//
//	@param form name of form file
//	@param dst destination file name
//
//	@return error
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

// Map 注入
//
//	@param i
//
//	@return inject.TypeMapper
func (c *Context) Map(i ...interface{}) inject.TypeMapper {
	return c.inj.Map(i...)
}

// MapTo 注入为某个接口
//
//	@param i 要注入的值（指针）
//	@param j 接口类型（指针）
//
//	@return inject.TypeMapper
func (c *Context) MapTo(i interface{}, j interface{}) inject.TypeMapper {
	return c.inj.MapTo(i, j)
}

// Apply 为一个结构注入带inject标签的Field
//
//	@param ctl 需要为指针
//
//	@return error
func (c *Context) Apply(ctl interface{}) error {
	return c.inj.Apply(ctl)
}

// Provide 提供已注入的实例
//
//	@param i 需要为指针
//
//	@return error
func (c *Context) Provide(i interface{}) error {
	return c.inj.Provide(i)
}
