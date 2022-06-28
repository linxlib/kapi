package controller

import (
	"fmt"
	"github.com/linxlib/kapi"
	"time"
)

type MyService interface {
	Test() string
}
type MyServiceImpl struct {
}

func (m *MyServiceImpl) Test() string {
	return "MyServiceImpl"
}

//Example 例子
//@ROUTE /api/v1/example
type Example struct {
}

func (e *Example) Before(context *kapi.Context) {
	context.Set("start", time.Now())
}

func (e *Example) After(context *kapi.Context) {
	i, _ := context.Get("start")
	t := i.(time.Time)
	fmt.Println(time.Now().Sub(t).Microseconds(), "us")
	context.Abort()
}

type MyReq struct {
	Page int `query:"page,default=1"`
	Size int `query:"size,default=15"`
}

//GetList 获取列表
//@GET /list
func (e *Example) GetList(
	c *kapi.Context,
	req *MyReq,
	svc MyService,
) {
	fmt.Println(req.Page, req.Size)
	fmt.Println(svc.Test())
	c.Success()
}

//@GET /test
func (e *Example) TestPure(c *kapi.Context) {
	c.SuccessExit()
}
