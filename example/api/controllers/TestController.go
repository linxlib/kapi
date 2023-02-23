package controllers

import (
	kapi "github.com/linxlib/kapi"
	sub "github.com/linxlib/kapi/example/api/controllers/subpackage"
	model2 "github.com/linxlib/kapi/example/api/model"
)

const (
	A = "Test"
	B = false
)

var (
	CC        = "CC"
	DD string = "PP"
	EE bool   = false
	FF bool
)

// TestController 测试
// @AUTH
type TestController struct{}

type MyRequest struct {
	A string `json:"a"`
	B int    `json:"b"`
}

type MyResult struct {
	A string
	B string
}

// CreateOne
// @POST /createone
// @DEPRECATED
// @RESP MyResult
func (t *TestController) CreateOne(c *kapi.Context, req *MyRequest) {

}

type MyRequest1 struct {
	Result sub.MyResult5
}
type MyResult1 struct{}

// GetList
// @GET /getlist
// @RESP MyResult1
func (t *TestController) GetList(c *kapi.Context, req *MyRequest1) (*MyResult1, error) {
	return nil, nil
}

// GetList2
// @POST /getlist2
// @RESP MyResult2
func (t *TestController) GetList2(c *kapi.Context, req MyRequest2) (*MyResult2, error) {
	return nil, nil
}

// GetList3 获取列表
// here is some description
// @POST /getlist3/:c
// @RESP model.MyResult3
func (t *TestController) GetList3(c *kapi.Context, req *model2.MyRequest3) (*model2.MyResult3, error) {
	return nil, nil
}

// GetList5
// @GET /getlist5/:c
// @RESP sub.MyResult5
func (t *TestController) GetList5(c *kapi.Context, req *model2.MyRequest3) (*sub.MyResult5, error) {
	return nil, nil
}

// GetList6
// @GET /getlist6/:c
// @RESP string
func (t *TestController) GetList6(c *kapi.Context, req *model2.MyRequest3) {
	return
}

type Name string

const (
	AA Name = "ijiji" //uuuu
	//BB iijhg
	BB Name = "sdds" ///hhhh
	OO Name = "ydgydf"
)
const (
	U1 Name = "edfed"
)

// GetList7
// @GET /getlist7/:c
// @RESP Name
func (t *TestController) GetList7(c *kapi.Context, req model2.MyRequest3) {
	return
}
