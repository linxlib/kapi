package controller

import "gitee.com/kirile/kapi"

//@TAG 测试标签
type Hello struct {
}

type TestReq struct {
	Page int `query:"page"` //页码
	Size int `query:"size"` //页数量
}

// @GET /hello/list
func (h *Hello) List(c *kapi.Context, req *TestReq) {
	list := make([]TestReq, 0)
	list = append(list, *req)
	count := int64(30)
	c.ListExit(count, list)

}
