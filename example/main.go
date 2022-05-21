package main

import (
	"github.com/linxlib/kapi"
	"github.com/linxlib/kapi/ast"
	"test_kapi/api/controller"
)

func main() {
	ast.AddImportFile("mime/multipart", "mime/multipart")
	k := kapi.New(func(option *kapi.Option) {
		// 默认读取config.toml 在这可以覆盖配置文件中的设置
		option.Server.Port = 8087
	})
	//此处解析路由和注册路由
	k.RegisterRouter(new(controller.Hello),
		new(controller.CompanyController),
		new(controller.UserController),
		new(controller.LoginController),
		new(controller.CategoryController),
	)

	k.Run()

}
