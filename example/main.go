package main

import (
	"gitee.com/kirile/kapi"
	"test_kapi/routers"
)

func main() {
	k := kapi.New(
		kapi.WithDebug(true),
		kapi.OutputDoc("测试"),

		kapi.OpenDoc(),
		kapi.Port(8081),
	)
	k.RegisterRouter(routers.Register)
	k.Run()

}
