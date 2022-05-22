package controller

import (
	"fmt"
	"github.com/linxlib/kapi"
	"mime/multipart"
	"test_kapi/api/model"
	"test_kapi/lib"
)

//CategoryController hhh
//@TAG hudhushd和上面这行同时存在时，这个优先 对了 注解后有空格的话，空格后的内容会被无视
//@AUTH Authorization 表示下面的每个方法都会需要Authorization这个Header Authorization是默认 所以也可以不写
type CategoryController struct {
}

type GetCategoryListReq struct {
	PageSize
}

//GetCategoryList k
//@GET /category/list
func (p *CategoryController) GetCategoryList(c *kapi.Context, req *GetCategoryListReq) {
	fmt.Println(req.PageSize)
}

type GetCategoryOneReq struct {
	ID int64 `query:"id"`
	//State State `query:"state"`

}

//GetCategoryOne k
//@GET /category
func (p *CategoryController) GetCategoryOne(c *kapi.Context, req *GetCategoryOneReq) (*GetCategoryListReq, error) {

	return nil, nil
}

type PostCategoryReq struct {
	File *multipart.FileHeader `form:"file"`
}

//PostCategory k
//@POST /category
//@RESP lib.User
func (p *CategoryController) PostCategory(c *kapi.Context, req *PostCategoryReq) {
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
//@GET /TestAnyMethod2
//@POST /TestAnyMethod1 但是上面的不能不写 ^_^
//@PUT /TestAnyMethode 只会以最下面这个为准 以后会更新 让每个都不一样
func (p *CategoryController) TestAnyMethod(c *kapi.Context, req *DelCategoryReq) {

}
