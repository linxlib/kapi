# kapi

#### 介绍
基于gin的扩展, 支持 //@POST 式注释路由, 解析结构体并生成swagger文档

#### 特性
1. 使用注释路由
2. 更简单的swagger文档生成方式


#### 快速开始
1. 在go module项目基础上创建routers目录, 并在其中的`router.go`创建
```go
func Register(b *kapi.KApi, r *gin.Engine) {
	g := r.Group("/api")
	{
		b.Register(g, new(controller.HelloController))
		// 这里也可以多次调用 b.Register 来注册
	}
}
```
2. main.go 中初始化
```go
        k := kapi.New(
		kapi.WithDebug(true), //指定是否开发模式
		kapi.OutputDoc("问卷系统"), //swagger文档名称
		//kapi.OpenDoc(), //运行时 在windows下打开浏览器访问swagger
		kapi.WithDomain(domain), //为swagger文档指定域名地址, 用于线上直接使用swagger调用接口(一般线上不会这么干)
		kapi.Port(httpPort)) //指定端口

	k.RegisterRouter(routers.Register) //注册
	k.Run() //运行
```
3. 创建 controller.HelloController , 并实现一个路由方法
```go
// World1 World1
// 一个Context参数的方法
// @GET /hello/list
func (h *HelloController) World1(c *kapi.Context) {
	c.SuccessExit()
}
```
4. 运行

# k
#### 介绍
这是本框架提供的命令行工具, 代码基本来自 `github.com/gogf/gf-cli`, 目前包含 安装、运行、编译三个部分， 后续会加入其它功能