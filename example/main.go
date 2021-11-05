package main

import (
	"gitee.com/kirile/kapi"
	"test_kapi/api/controller"
)

func main() {
	k := kapi.New(func(option *kapi.Option) {
		option.SetIsDebug().
			SetDocName("测试").
			SetOpenDocInBrowser().
			SetApiBasePath("/api2").
			SetPort(8081)
	})
	k.RegisterRouter(new(controller.Hello))
	k.Run()

}
