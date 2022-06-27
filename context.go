package kapi

import (
	"github.com/gin-gonic/gin"
	"github.com/linxlib/conv"
	"github.com/linxlib/kapi/inject"
	"mime/multipart"
	"reflect"
	"strings"
)

type Handler interface{}

// Context Wrapping gin context to custom context
type Context struct { // 包装gin的上下文到自定义context
	inject.Injector
	ctx     *gin.Context
	handler Handler
}

// newContext Create a new custom context
func newContext(c *gin.Context) *Context { // 新建一个自定义context
	return &Context{
		Injector: inject.New(),
		ctx:      c,
	}
}

// NewAPIFunc default of custom handlefunc
func NewAPIFunc(c *gin.Context) interface{} {
	return newContext(c)
}

var KAPIEXIT = "kapiexit"

func (c *Context) Exit() {
	panic(KAPIEXIT)
}

func (c *Context) Method() string {
	return c.ctx.Request.Method
}

func (c *Context) Get(key string) (interface{}, bool) {
	return c.ctx.Get(key)
}
func (c *Context) ClientIP() string {
	return c.ctx.ClientIP()
}
func (c *Context) Header(key string, value string) {
	c.ctx.Header(key, value)
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	return c.ctx.FormFile(name)
}
func (c *Context) Data(code int, contentType string, data []byte) {
	c.ctx.Data(code, contentType, data)
}

func (c *Context) GetQueryString(key string) string {
	tmp, b := c.ctx.GetQuery(key)
	if b {
		return tmp
	} else {
		return ""
	}
}

func (c *Context) GetQueryInt(key string, def ...int) int {
	tmp, b := c.ctx.GetQuery(key)
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
	tmp, b := c.ctx.GetQuery(key)
	if b {
		return conv.Int64(tmp)
	} else {
		return -1
	}
}

func (c *Context) GetQueryUInt64(key string) uint64 {
	tmp, b := c.ctx.GetQuery(key)
	if b {
		return conv.Uint64(tmp)
	} else {
		return 0
	}
}

func (c *Context) GetQueryUInt(key string) uint {
	tmp, b := c.ctx.GetQuery(key)
	if b {
		return conv.Uint(tmp)
	} else {
		return 0
	}
}

func (c *Context) GetQueryBool(key string) bool {
	tmp, b := c.ctx.GetQuery(key)
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

func (c *Context) DefaultPostForm(form string, defaultValue string) string {
	return c.ctx.DefaultPostForm(form, defaultValue)
}

func (c *Context) SaveFile(form string, dst string) error {
	f, err := c.ctx.FormFile(form)
	if err != nil {
		return err
	}
	err = c.ctx.SaveUploadedFile(f, dst)
	if err != nil {
		return err
	}
	return nil
}

// RemoteAddr returns more real IP address.
func (c *Context) RemoteAddr() string {
	addr := c.ctx.Request.Header.Get("X-Real-IP")
	if len(addr) == 0 {
		addr = c.ctx.Request.Header.Get("X-Forwarded-For")
		if addr == "" {
			addr = c.ctx.Request.RemoteAddr
			if i := strings.LastIndex(addr, ":"); i > -1 {
				addr = addr[:i]
			}
		}
	}
	return addr
}

func (c *Context) run() ([]reflect.Value, error) {
	return c.Invoke(c.handler)
}
