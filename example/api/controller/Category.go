package controller

import (
	"github.com/linxlib/kapi"
	"test_kapi/api/model"
	"test_kapi/lib"
)

//@AUTH Authorzation
type CategoryController struct {
}

type GetCategoryListReq struct {
	PageSize
}

//GetCategoryList
//@GET /category/list
func (p *CategoryController) GetCategoryList(c *kapi.Context, req *GetCategoryListReq) {

}

type GetCategoryOneReq struct {
	ID int64 `query:"id"`
}

//GetCategoryOne
//@GET /category
func (p *CategoryController) GetCategoryOne(c *kapi.Context, req *GetCategoryOneReq) (*GetCategoryListReq, error) {

	return nil, nil
}

//PostCategory
//@POST /category
//@RESP lib.User
func (p *CategoryController) PostCategory(c *kapi.Context, req *model.Category) {
	var u = lib.User{}
	c.DataExit(u)
}

//PutCategory
//@PUT /category
func (p *CategoryController) PutCategory(c *kapi.Context, req *model.Category) {

}

type DelCategoryReq struct {
	ID int64 `path:"id"`
}

//DelCategory
//@DELETE /category/:id
func (p *CategoryController) DelCategory(c *kapi.Context, req *DelCategoryReq) {

}

//TestAnyMethod
//@GET /TestAnyMethod
//@POST /TestAnyMethod
//@PUT /TestAnyMethod
func (p *CategoryController) TestAnyMethod(c *kapi.Context, req *DelCategoryReq) {

}
