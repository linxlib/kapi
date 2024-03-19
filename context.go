package kapi

import (
	"github.com/gin-gonic/gin"
	"github.com/linxlib/conv"
	"github.com/linxlib/inject"
	"strings"
)

// Context KApi Context
type Context struct {
	*gin.Context
	inj           inject.Injector
	ResultBuilder IResultBuilder `inject:""`
}

// newContext create a new custom context
func newContext(c *gin.Context, parent ...inject.Injector) *Context {
	if len(parent) > 0 {
		cc := &Context{
			Context: c,
			inj:     inject.New(),
		}
		cc.inj.SetParent(parent[0])
		err := cc.inj.Apply(cc)

		if err != nil {
			cc.ResultBuilder = &DefaultResultBuilder{}
		}
		return cc
	} else {
		cc := &Context{
			Context: c,
			inj:     inject.New(),
		}
		err := cc.inj.Apply(cc)
		if err != nil {
			cc.ResultBuilder = &DefaultResultBuilder{}
		}
		return cc
	}
}

// Method return HTTP method
//
//	@return string
func (c *Context) Method() string {
	return c.Request.Method
}

// GetQueryString return the string value of query param
//
//	@param key
//
//	@return string
func (c *Context) GetQueryString(key string, def ...string) string {
	tmp, b := c.GetQuery(key)
	if b {
		return tmp
	} else {
		if len(def) > 0 {
			return def[0]
		}
		return ""
	}
}

// GetFormString return the string value of form param
//
//	@param key
//
//	@return string
func (c *Context) GetFormString(key string, def ...string) string {
	tmp, b := c.GetPostForm(key)
	if b {
		return tmp
	} else {
		if len(def) > 0 {
			return def[0]
		}
		return ""
	}
}

// GetFormInt return the int value of form param
//
//	@param key
//
//	@return string
func (c *Context) GetFormInt(key string, def ...int) int {
	tmp, b := c.GetPostForm(key)
	if b {
		return conv.Int(tmp)
	} else {
		if len(def) > 0 {
			return def[0]
		}
		return -1
	}
}

// GetFormInt64 return the int64 value of form param
//
//	@param key
//
//	@return string
func (c *Context) GetFormInt64(key string, def ...int64) int64 {
	tmp, b := c.GetPostForm(key)
	if b {
		return conv.Int64(tmp)
	} else {
		if len(def) > 0 {
			return def[0]
		}
		return -1
	}
}

// GetFormUInt return the uint value of form param
//
//	@param key
//
//	@return string
func (c *Context) GetFormUInt(key string, def ...uint) uint {
	tmp, b := c.GetPostForm(key)
	if b {
		return conv.Uint(tmp)
	} else {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
}

// GetFormFloat32 return the float32 value of query param
//
//	@param key
//	@param def
//
//	@return string
func (c *Context) GetFormFloat32(key string, def ...float32) float32 {
	tmp, b := c.GetPostForm(key)
	if b {
		return conv.Float32(tmp)
	} else {
		if len(def) > 0 {
			return def[0]
		}
		return 0.0
	}
}

// GetFormFloat64 return the float64 value of form param
//
//	@param key
//	@param def
//
//	@return string
func (c *Context) GetFormFloat64(key string, def ...float64) float64 {
	tmp, b := c.GetPostForm(key)
	if b {
		return conv.Float64(tmp)
	} else {
		if len(def) > 0 {
			return def[0]
		}
		return 0.0
	}
}

// GetQueryInt return the int value of query param
//
//	@param key
//	@param def
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

// GetQueryInt64 return the int64 value of query param
//
//	@param key
//	@param def
//
//	@return int64
func (c *Context) GetQueryInt64(key string, def ...int64) int64 {
	tmp, b := c.GetQuery(key)
	if b {
		return conv.Int64(tmp)
	} else {
		if len(def) > 0 {
			return def[0]
		}
		return -1
	}
}

// GetQueryUInt64 return the uint64 value of query param
//
//	@param key
//	@param def
//
//	@return uint64
func (c *Context) GetQueryUInt64(key string, def ...uint64) uint64 {
	tmp, b := c.GetQuery(key)
	if b {
		return conv.Uint64(tmp)
	} else {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
}

// GetQueryUInt return the uint value of query param
//
//	@param key
//	@param def
//
//	@return uint
func (c *Context) GetQueryUInt(key string, def ...uint) uint {
	tmp, b := c.GetQuery(key)
	if b {
		return conv.Uint(tmp)
	} else {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
}

// GetQueryBool return the boolean value of query param
//
//	@param key
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

// GetPageSize return the values of "page"-1 and "size" query params
//
//	@param def
//	def[0] can be the default value of "page" param and def[1] can be the default value of "size" param
//
//	@return int page-1
//	@return int size
func (c *Context) GetPageSize(def ...int) (int, int) {
	defPage := 1
	defSize := 10
	if len(def) > 1 {
		defPage = def[0]
		defSize = def[1]
	}
	page := c.GetQueryInt("page", defPage)
	size := c.GetQueryInt("size", defSize)
	return conv.Int(page) - 1, conv.Int(size)
}

// OfFormFile 获取Form File的文件名和大小
//
//	@param form
//
//	@return string
//	@return int64
func (c *Context) OfFormFile(form string) (string, int64) {
	f, err := c.FormFile(form)
	if err != nil {
		return "", 0
	}
	return f.Filename, f.Size
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
	if dst == "" {
		dst = f.Filename
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

// Map an instance
//
//	@param i
//
//	@return inject.TypeMapper
func (c *Context) Map(i ...interface{}) inject.TypeMapper {
	return c.inj.Map(i...)
}

// MapTo map instance to interface
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

// File returns a file
//
//	@param fileName
//	@param fileData
func (c *Context) File(fileName string, fileData []byte) {
	c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	c.Data(200, "application/octet-stream", fileData)
}
