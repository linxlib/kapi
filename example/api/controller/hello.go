package controller

import (
	"gitee.com/kirile/kapi"
)

//@TAG 测试标签
//@ROUTE /hello2
type Hello struct {
}

type PageSize struct {
	Page int `query:"page" default:"1"`
	Size int `query:"size" default:"15"`
}

func (p *PageSize) GetLimit() (int, int) {
	return p.Size, (p.Page - 1) * p.Size
}

type TestReq struct {
	PageSize
	Name string `query:"name"` //名称查找
}

//List 获取列表
// @GET /list
func (h *Hello) List(c *kapi.Context, req *TestReq) {
	list := make([]TestReq, 0)
	list = append(list, *req)
	count := int64(30)
	c.ListExit(count, list)

}

// @GET /hello/list2
func (h *Hello) List2(c *kapi.Context, req *TestReq) ([]TestReq, error) {
	list := make([]TestReq, 0)
	list = append(list, *req)
	//count := int64(30)
	return list, nil
}

func init() {
	kapi.RegisterController(new(Hello))
}
