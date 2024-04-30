package controllers

import (
	"github.com/linxlib/kapi"
)

// HelloController
// @ROUTE /hello
type HelloController struct{}

type MyBody struct{}
type MyResult struct{}

// World1
// @GET /world1
// @RESP MyResult
func (h *HelloController) World1(c *kapi.Context, req *MyBody) {

}
