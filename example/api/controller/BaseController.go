package controller

import "github.com/linxlib/kapi"

type BaseController struct {
	MyCustomData string
}

func (b *BaseController) Before(req *kapi.InterceptorContext) bool {
	b.MyCustomData = "Hello World"
	return true
}

func (b *BaseController) After(req *kapi.InterceptorContext) bool {
	return true
}
