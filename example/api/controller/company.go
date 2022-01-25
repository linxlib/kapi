package controller

import (
	"gitee.com/kirile/kapi"
	"test_kapi/api/model"
)

type CompanyController struct {
}

type GetCompanyListReq struct {
	PageSize
}

//GetCompanyList
//@GET /company/list
func (p *CompanyController) GetCompanyList(c *kapi.Context, req *GetCompanyListReq) {

}

type GetCompanyOneReq struct {
	ID int64 `query:"id"`
}

//GetCompanyOne
//@GET /company
func (p *CompanyController) GetCompanyOne(c *kapi.Context, req *GetCompanyOneReq) {

}

//PostCompany
//@POST /company
func (p *CompanyController) PostCompany(c *kapi.Context, req *model.Company) {

}

//PutCompany
//@PUT /company
func (p *CompanyController) PutCompany(c *kapi.Context, req *model.Company) {

}

type DelCompanyReq struct {
	ID int64 `path:"id"`
}

//DelCompany
//@DELETE /company/:id
func (p *CompanyController) DelCompany(c *kapi.Context, req *DelCompanyReq) {

}
