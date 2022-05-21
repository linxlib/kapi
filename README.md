# kapi

## PS
目前正在从gitee转移到github...., 慢慢更新中

## 介绍
基于gin的扩展, 支持 //@POST 式注释路由, 解析结构体并生成swagger文档


注意: 还有很多问题, 不要在生产环境使用, 会很难受的

## 特性
1. 使用注释路由
2. 更简单的swagger文档生成方式


## 注意
由于使用了go 1.16新增的embed特性, 因此只能在1.16以上版本使用

## 快速开始

1. 到 `https://github.com/linxlib/k` 安装cli
2. 安装后执行`go mod init {module_name}` 然后 `k init`, 自动创建 config.toml build.toml main.go 等
3. 
```go
   k := kapi.New(func(option *kapi.Option) {
	   // 这里也可以为空 已经从配置文件中进行读取
	   // 也可以在config.toml中进行修改, 这里写的话将覆盖配置文件中的设置
        option.SetNeedDoc(true) //是否需要生成swagger文档
        option.SetDocName("系统")
        option.SetDocDescription("系统api")
        option.SetIsDebug(true)
        option.SetPort(3080)
        option.SetDocVersion("")
        //option.SetApiBasePath("/")
       //option.SetDocDomain("http://example.com") // 使可以支持在线上用域名访问文档, 本地开发时默认为http://局域网ip+端口/swagger/
        option.SetRedirectToDocWhenAccessRoot(true) //访问 / 时跳转到文档
        option.SetStaticDirs("html", "h5")  //可以设置多个static目录
    })
    // 注册路由
    k.RegisterRouter(new(controller.BannerController),
        new(controller.AssetController),
        new(controller.CategoryController),
        new(controller.UserController),
        new(controller.SysConfigController))
	k.Run() //运行
```
3. 创建 controller.HelloController , 并实现一个路由方法
```go
//HelloController ...
//@TAG HelloController
//@AUTH Authorization
//@ROUTE /api
type HelloController struct {
	BaseAuthController
}

type PageSize struct {
	Page int `query:"page" default:"1"`
	Size int `query:"size" default:"15"`
}

func (ps *PageSize) GetLimit() (int,int) {
	return ps.Size,(ps.Page-1)*ps.Size
}

type HelloWorld1Req struct {
	PageSize
	Name string `query:"name" binding:"required"`
	Authorization string `header:"Authorization"`
	
}
// World1 World1
// @GET /hello/list
func (h *HelloController) World1(c *kapi.Context,req *HelloWorld1Req) {
	c.SuccessExit()
}
```
其他路由和swagger文档相关请到example中查看

4. 运行 `go run main.go` 或者 `k run`

具体可查看example文件夹


## 支持一些奇怪的特性 🐶

- `//@TAG 分类` 在struct上增加, 可以指定在swagger文档中的标签, 默认为struct的名字, **目前不加@TAG注释, 则默认使用struct自己的注释**
- 一个方法`List`上如果有这样的注释 `//List 获取列表` 那么`获取列表` 将作为一个路由的Summary显示在swagger文档里
- `//@AUTH Authorization` 在struct上增加, 可以为该struct的每个方法的请求参数加上一个Header请求头, 其中 `Authorization` 可以不要, 默认是 `Authorization`. 
这个需要配合 `BaseAuthController`来对各个方法进行鉴权
- `//@ROUTE /banner` 为该struct下的方法增加一个路由地址的前缀, 会拼接起来. 例如 struct上有`//@ROUTE /banner`, 其下方的方法`//@GET /list` 则实际的路由为 `GET /banner/list`
- 请求的参数可以使用类似继承的方式来声明参数. 
```go
type PageSize struct {
    Page int `query:"page" default:"1"`
    Size int `query:"size" default:"15"`
}

func (p *PageSize) GetLimit() (int, int) {
    return p.Size, (p.Page - 1) * p.Size
}
type GetBannerListReq struct {
	PageSize
}
```
- 请求参数的struct支持多种tag, `query` `path` `header` `json` `default` 和 `binding`, 这个是基于gin的bind实现的, 
由于不同类型的参数混在一起, 因此这里可能需要优化性能

- `kapi.Context`包含一些Exit方法, 可以不用return直接跳出流程, 这是通过panic实现的, 当然如果方法使用了返回值, 就不能用这个方式了
- 实现了 kapi.Interceptor 的中间件, 可以存储一些上下文数据, 比如 当前用户 CurrentUser *model.User, 无需使用Context的相关方法
- `kapi.RegisterFuncGetResult` 可以修改默认的返回json的结构, 为自带的 `*Exit`系方法自定义返回
- since v0.3.2 `//@RESP model.User` 可以注明方法返回结构体，用于没有返回值的方法在文档中显示返回类型
- since v0.3.2 可以在方法上指定多个HTTP METHOD 例如可以写多行 @GET @POST @PUT， 不过请求路径会以最后一个为准

## 部署
`k build`后 `./bin/版本/系统_架构/`目录下的文件即为全部, 如果是自行编译, 则需要同时拷贝swagger.json和gen.gob以及config.toml.

当前主分支, 在 `k build -g` 后会覆盖 `swagger.json` 和 `gen.gob` ,`config.toml`则仅当输出目录下不存在`config.toml`时才会拷贝


## TODO

- [x] 修改为func方式配置kapi
- [x] controller 增加 //@ROUTE 用于标记整个controller的path
- [x] 增加@AUTH标记支持, 用于设置传输token的header名, 可以放在controller上
- [x] 增加静态目录配置, Context增加 SaveFile
- [x] 在issue中进行任务处理
- [x] 加入二维数组支持
- [x] 请求参数可以使用类似继承的方式来重用结构
- [x] 配置文件配置服务 (需要配合 k init)
- [x] 增加命令行参数用于仅生成路由和文档, 实现编译前无需运行即可更新文档
- [x] 优化ast包解析, 减少循环 (目前通过增加map来缓存需要的数据, 重复的对象不会多次遍历ast树)
- [x] k cli 加入项目判断, 使其可用于其他纯go项目的编译
- [x] 重构ast解析部分，提升效率
- [ ] 部分功能提取为单独包
- [ ] 拦截器实现多个顺序执行机制（栈）
- [ ] 加入枚举支持
- [ ] [RapiDoc](https://github.com/rapi-doc/RapiDoc)
- [x] 增加一个注解用于注释返回结构类型, 特指只有一个Context的方法
- [ ] k编译打包时增加进度显示, 优化打包速度
- [ ] 精简引用包，减小体积
- [ ] 加入markdown形式的文档
  
## 感谢

`https://github.com/xxjwxc/ginrpc` 大部分代码参考该项目, 例如代码解析\文档生成, 由于我需要一个方便单纯写写接口, 快速生成文档, 该项目无法满足, 
而且也很难在基础上进行PR(改动较大, 并可能无法适应比较大众化的框架需求), 才有了魔改一顿的想法

# k
#### 介绍
这是本框架提供的命令行工具, 代码基本来自 `github.com/gogf/gf-cli`, 目前包含 安装、运行、编译三个部分， 后续会加入其它功能.

使用kapi进行开发, 建议同时使用k-cli, 由于kapi的swagger文档以及路由注册需要在开发环境运行后才会生成, 使用go自带的编译可能无法正常使用文档和注册路由

**目前移动到 https://github.com/linxlib/k** 仓库


