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

- [ ] 加入枚举支持

- [x] 加入二维数组支持 

- [x] 请求参数可以使用类似继承的方式来重用结构
- 
- [ ] 配置文件配置服务
- 
- [ ] 增加命令行参数用于仅生成路由和文档, 实现编译前无需运行即可更新文档
  


# k
#### 介绍
这是本框架提供的命令行工具, 代码基本来自 `github.com/gogf/gf-cli`, 目前包含 安装、运行、编译三个部分， 后续会加入其它功能

**目前移动到https://gitee.com/kirile/k-cli仓库**


# app

#### 介绍
app包, 包装了xorm mysql 和 redis的初始化连接以及一些简单的封装方法. 还包括了一个toml读取的辅助

**目前移动到https://gitee.com/kirile/kapp**