package kapi

import (
	"reflect"
)

type ContextInvoker func(ctx *Context)

func (invoke ContextInvoker) Invoke(params []interface{}) ([]reflect.Value, error) {
	invoke(params[0].(*Context))
	return nil, nil
}
