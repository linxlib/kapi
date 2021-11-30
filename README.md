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

1. 到 `https://gitee.com/kirile/k-cli` 安装cli
2. 安装后执行`go mod init {module_name}` 然后 `k init`, 自动创建 config.toml build.toml main.go 等
3. 
```go
   k := kapi.New(func(option *kapi.Option) {
	   // 也可以在config.toml中进行修改, 这里写的话将覆盖配置文件中的设置
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
// World1 World1
// 一个Context参数的方法
// @GET /hello/list
func (h *HelloController) World1(c *kapi.Context) {
	c.SuccessExit()
}
```
其他路由和swagger文档相关请到example中查看

4. 运行 `go run main.go` 或者 `k run`

具体可查看example文件夹


## 支持一些奇怪的特性 🐶

- `//@TAG 分类` 在struct上增加, 可以指定在swagger文档中的标签, 默认为struct的名字
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
## 部署
`k build`后 `./bin/版本/系统_架构/`目录下的文件即为全部, 如果是自行编译, 则需要同时拷贝swagger.json和gen.gob以及config.toml

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
- [ ] 加入枚举支持
- [ ] k cli 加入项目判断, 使其可用于其他纯go项目的编译
  
## 感谢

`https://github.com/xxjwxc/ginrpc` 大部分代码参考该项目, 例如代码解析\文档生成, 由于我需要一个方便单纯写写接口, 快速生成文档, 该项目无法满足, 
而且也很难在基础上进行PR(改动较大, 并可能无法适应比较大众化的框架需求), 才有了魔改一顿的想法

# k
#### 介绍
这是本框架提供的命令行工具, 代码基本来自 `github.com/gogf/gf-cli`, 目前包含 安装、运行、编译三个部分， 后续会加入其它功能

**目前移动到https://gitee.com/kirile/k-cli仓库**


# app

#### 介绍
app包, 包装了xorm mysql 和 redis的初始化连接以及一些简单的封装方法. 还包括了一个toml读取的辅助

**目前移动到https://gitee.com/kirile/kapp**