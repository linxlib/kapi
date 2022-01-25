package controller

import (
	"fmt"
	"gitee.com/kirile/kapi"
	"test_kapi/api/model"
)

type LoginController struct {
	BaseController
}

type GetLoginListReq struct {
	PageSize
}

//GetLoginList
//@GET /login/list
func (p *LoginController) GetLoginList(c *kapi.Context, req *GetLoginListReq) {
	fmt.Println(p.MyCustomData)
}

type GetLoginOneReq struct {
	ID int64 `query:"id"`
}

//GetLoginOne
//@GET /login
func (p *LoginController) GetLoginOne(c *kapi.Context, req *GetLoginOneReq) {

}

//PostLogin
//@POST /login
func (p *LoginController) PostLogin(c *kapi.Context, req *model.User) {

}

//PutLogin
//@PUT /login
func (p *LoginController) PutLogin(c *kapi.Context, req *model.User) {

}

type DelLoginReq struct {
	ID int64 `path:"id"`
}

//DelLogin
//@DELETE /login/:id
func (p *LoginController) DelLogin(c *kapi.Context, req *DelLoginReq) {

}
