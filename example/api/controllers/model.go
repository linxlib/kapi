package controllers

import model2 "github.com/linxlib/kapi/example/api/model"

type MyRequest2 struct {
	Request3 model2.MyRequest3 `json:"request3"`
	Names    Name              `query:"names" v:"required"`
}
type MyResult2 struct{}
