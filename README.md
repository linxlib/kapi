# kapi

## 介绍
基于gin的扩展, 支持 //@POST 式注释路由, 解析结构体并生成swagger文档

## 特性
1. 使用注释路由
2. 更简单的swagger文档生成方式


## 注意
由于使用了go 1.16新增的embed特性, 因此只能在1.16以上版本使用

## 快速开始
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

具体可查看example文件夹

## TODO

- [ ] 增加一种注册controller的方式，在包的init中注册
- [x] 修改为func方式配置kapi
- [ ] 换个地方生成gen_code.go  
- [ ] 注册成gRPC服务
- [x] controller 增加 //@ROUTE 用于标记整个controller的path


# k
#### 介绍
这是本框架提供的命令行工具, 代码基本来自 `github.com/gogf/gf-cli`, 目前包含 安装、运行、编译三个部分， 后续会加入其它功能

#### 安装&更新

`go get -u gitee.com/kirile/kapi/k`

#### 命令

- `k run` 运行, 监听源码变动, 重新运行
- `k` 不带参数运行则默认为安装可执行程序到系统 PATH目录中
- `k build` 依赖项目根目录下的 config.toml 文件
```toml
# 编译配置
[k]
  name = "api_base" # 编译的可执行文件名
  version = "1.0.0" # 版本
  arch = "amd64"  # 平台
  system = "darwin" # 系统
  path = "./bin" # 输出目录
```
使用k编译的kapi程序, 运行时将输出相关日志信息, 例如
```
--------------------------------------------
    _/    _/    _/_/    _/_/_/    _/_/_/
   _/  _/    _/    _/  _/    _/    _/
  _/_/      _/_/_/_/  _/_/_/      _/
 _/  _/    _/    _/  _/          _/
_/    _/  _/    _/  _/        _/_/_/

 Version:   1.0.0/go1.16.6
 OS/Arch:   windows/amd64
 BuiltTime: 2021-08-20T11:31:08
-------------------------------------------- 
```
如果config.toml不存在, 则只会在./bin目录下输出当前系统的可执行程序(相当于 go build main.go)


# app

#### 介绍
app包, 包装了xorm mysql 和 redis的初始化连接以及一些简单的封装方法. 还包括了一个toml读取的辅助

目前此包单独发布，app/1.0.0