package routers

import (
	"gitee.com/kirile/kapi"
	"github.com/gin-gonic/gin"
	"test_kapi/api/controller"
)

func Register(b *kapi.KApi, r *gin.Engine) {
	g := r.Group("/api")
	{
		b.Register(g, new(controller.Hello))
	}
}
