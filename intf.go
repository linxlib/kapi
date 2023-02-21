package kapi

// Interceptor implement this to intercept controller method
type Interceptor interface {
	Before(*Context)
	After(*Context)
}

type HeaderAuth interface {
	HeaderAuth(c *Context)
}

type BeforeBind interface {
	BeforeBind(c *Context)
}

type AfterBind interface {
	AfterBind(c *Context)
}

type BeforeCall interface {
	BeforeCall(c *Context)
}

type AfterCall interface {
	AfterCall(c *Context)
}

type OnPanic interface {
	OnPanic(c *Context, err interface{})
}

type OnError interface {
	OnError(c *Context, err error)
}

type OnValidationError interface {
	OnValidationError(c *Context, err error)
}

type OnUnmarshalError interface {
	OnUnmarshalError(c *Context, err error)
}
