package controller

import (
	"gitee.com/kirile/kapi"
	"test_kapi/api/model"
)

type UserController struct {
}

type GetUserListReq struct {
	PageSize
}

//GetUserList
//@GET /user/list
func (p *UserController) GetUserList(c *kapi.Context, req *GetUserListReq) {

}

type GetUserOneReq struct {
	ID int64 `query:"id"`
}

//GetUserOne
//@GET /user
func (p *UserController) GetUserOne(c *kapi.Context, req *GetUserOneReq) {

}

//PostUser
//@POST /user
func (p *UserController) PostUser(c *kapi.Context, req *model.User) {

}

//PutUser
//@PUT /user
func (p *UserController) PutUser(c *kapi.Context, req *model.User) {

}

type DelUserReq struct {
	ID int64 `path:"id"`
}

//DelUser
//@DELETE /user/:id
func (p *UserController) DelUser(c *kapi.Context, req *DelUserReq) {

}
