# kapi

## 介绍
基于gin的扩展, 支持 //@POST 式注释路由, 解析结构体并生成swagger文档



注意: 还有很多问题, 不要在生产环境使用, 会很难受的

## 特性
1. 使用注释路由
2. 更简单的swagger文档生成方式


## 注意
由于使用了go 1.16新增的embed特性, 因此只能在1.16以上版本使用

## 快速开始

1. main.go 中初始化
```go
   k := kapi.New(func(option *kapi.Option) {
        option.SetNeedDoc(true)
        option.SetDocName("系统")
        option.SetDocDescription("系统api")
        option.SetIsDebug(true)
        option.SetPort(3080)
        option.SetDocVersion("")
        //option.SetApiBasePath("/")
       //option.SetDocDomain("http://example.com")
        option.SetRedirectToDocWhenAccessRoot(true)
        option.SetStaticDir("asset")
    })

    k.RegisterRouter(new(controller.BannerController),
        new(controller.AssetController),
        new(controller.CategoryController),
        new(controller.UserController),
        new(controller.SysConfigController))
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

- [x] 修改为func方式配置kapi

- [x] controller 增加 //@ROUTE 用于标记整个controller的path

- [x] 增加@AUTH标记支持, 用于设置传输token的header名, 可以放在controller上

- [x] 增加静态目录配置, Context增加 SaveFile 

- [ ] 在issue中进行任务处理

  


# k
#### 介绍
这是本框架提供的命令行工具, 代码基本来自 `github.com/gogf/gf-cli`, 目前包含 安装、运行、编译三个部分， 后续会加入其它功能

#### 安装&更新

`go get -u gitee.com/kirile/kapi/k`

#### 命令

- `k init` 用于在`go.mod`存在的情况下为你生成`build.toml` `config.toml`  `main.go` 并执行`go mod tidy`

- `k run` 运行, 监听源码变动, 重新运行
- `k` 不带参数运行则默认为安装可执行程序到系统 PATH目录中
- `k build` 依赖项目根目录下的 build.toml 文件, **目前版本会在构建后一起拷贝 gen.gob  swagger.json 和配置文件到构建目录(配置文件已存在则不覆盖)**
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
如果build.toml不存在, 则只会在./bin目录下输出当前系统的可执行程序(相当于 go build main.go), 可以输入`k init` 生成这个文件


# app

#### 介绍
app包, 包装了xorm mysql 和 redis的初始化连接以及一些简单的封装方法. 还包括了一个toml读取的辅助

目前此包单独发布，app/1.0.0