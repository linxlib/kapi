package main

import (
	"github.com/linxlib/kapi"
	"test_kapi/core/api/v1/controller"
)

type MyResultBuilder struct {
	kapi.DefaultResultBuilder
}

type myMessageBody struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (m MyResultBuilder) OnSuccess(msg string, data any) (statusCode int, result any) {
	return 200, myMessageBody{
		Code: 200,
		Msg:  msg,
		Data: data,
	}
}

func (m MyResultBuilder) OnData(msg string, count int64, data any) (statusCode int, result any) {
	return 200, myMessageBody{
		Code: 200,
		Msg:  msg,
		Data: data,
	}
}

func (m MyResultBuilder) OnFail(msg string, data any) (statusCode int, result any) {
	return 200, myMessageBody{
		Code: 400,
		Msg:  msg,
		Data: data,
	}
}

func (m MyResultBuilder) OnError(msg string, err error) (statusCode int, result any) {
	return 200, myMessageBody{
		Code: 500,
		Msg:  msg,
		Data: err,
	}
}

func (m MyResultBuilder) OnErrorDetail(msg string, err any) (statusCode int, result any) {
	return 200, myMessageBody{
		Code: 500,
		Msg:  msg,
		Data: err,
	}
}

func (m MyResultBuilder) OnUnAuthed(msg string) (statusCode int, result any) {
	return 200, myMessageBody{
		Code: 401,
		Msg:  msg,
		Data: nil,
	}
}

func (m MyResultBuilder) OnNoPermission(msg string) (statusCode int, result any) {
	return 200, myMessageBody{
		Code: 403,
		Msg:  msg,
		Data: nil,
	}
}

func (m MyResultBuilder) OnNotFound(msg string) (statusCode int, result any) {
	return 200, myMessageBody{
		Code: 404,
		Msg:  msg,
		Data: nil,
	}
}

func main() {
	k := kapi.New(func(option *kapi.Option) {
		// 默认读取config.toml 在这可以覆盖配置文件中的设置
		option.Server.Port = 8087
		option.Server.Cors.AllowAllOrigins = false
	})
	k.RegisterResultBuilder(new(MyResultBuilder))
	// k := kapi.New() 也可以这样只使用配置文件进行配置
	//此处解析路由和注册路由
	k.RegisterRouter(new(controller.Example))
	k.MapTo(&controller.MyServiceImpl{}, (*controller.MyService)(nil))
	k.Run()
}
