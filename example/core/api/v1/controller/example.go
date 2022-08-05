package controller

import (
	"fmt"
	"github.com/linxlib/kapi"
	"strings"
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

type Base struct {
	CustomData string
}

func (b *Base) HeaderAuth(c *kapi.Context) {
	if strings.Contains(c.Request.RequestURI, "list") {
		b.CustomData = time.Now().Format(time.RFC3339)
	}
}

//Example 例子
//@ROUTE /api/v1/example
type Example struct {
	Base
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
	fmt.Println(e.CustomData)
	c.Success()
}

//@GET /test
func (e *Example) TestPure(c *kapi.Context) {
	a, _ := c.Get("123456")
	c.SuccessExit(a)
}
