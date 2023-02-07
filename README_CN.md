# KAPI

## 介绍
基于gin的扩展, 支持 //@METHOD 式注释路由, 解析结构体并生成swagger文档

最终目标: 敲几下命令，马上可以开始写HTTP接口


注意: 还有很多问题, 不要在生产环境使用, 会很难受的

## 特性
1. 通过注解的方式标注和注册路由
2. 使用类似Asp .Net Core的Controller开发形式
3. 使用 `https://github.com/linxlib/inject` 来进行依赖注入
4. 无需写一堆注释来生成swagger文档
5. 内置一个改造过得swagger界面（可复制请求path）


## 版本要求
由于增加了简单泛型使用，需要golang > 1.18


## 快速开始



## 注解支持

注解以 //@XXX parameter...  的方式设置 

| 注解 |位置|说明|  默认  | 参数可空 | 注解可选 |
|:--:|:----:| :----: |:----:|:--:|----|
| @TAG  | Struct | 指定在swagger文档中的标签 |注释|  | √ |
| @AUTH | Struct | 指定认证使用的Header名 |Authorization| √ | √ |
| @ROUTE | Struct | 指定Controller的路由前缀 |None|  | √ |
| @HTTPMETHOD | Method | 指定方法被注册成的路由 ||  |  |
| @RESP | Method | 指定方法的返回值的类型，生成文档中的Model ||  |  |



## 请求和响应的自动解析和文档生成

```go
// @GET /test
func (e *Example) TestPure(c *kapi.Context) {
	a, _ := c.Get("123456")
	c.SuccessExit(a)
}
```

可改写为

```go
// @GET /test
func (e *Example) TestPure(c *kapi.Context,req *MyRequest) {
	a, _ := c.Get("123456")
	c.SuccessExit(a)
}
```

kapi将自动解析MyRequest，并在swagger文档中体现

```go
type MyRequest struct {
  Page int `query:"page,default=1"` //注释， 可以设置参数默认值（默认暂时还只是在swagger中有效，实际不传参数时无效）
	Size int `query:"size,default=15"` // 指定该参数为从query中绑定
  MyHeader string `header:"MyHeader"`
  MyBodyParam1 string `json:"myBodyParam1"`
  MyBodyParam2 int `json:"myBodyParam2"`
  MyPathParam int `path:"id" binding:"required"` //设置参数为必选，swagger中会展示红色*号
  MyFormParam string `form:"myForm"` 
}
```

也可以这样

```go
type Form struct {
	FileHeader multipart.FileHeader `form:"file"` //这样swagger中就会展示一个选择文件上传的参数
}
```

可以使用 @RESP 来标注响应body的Model， 或者使用第二种写法

```go
type MyResult struct {
	Name string `json:"name"`
}

// @GET /test
// @RESP MyResult
func (e *Example) TestPure(c *kapi.Context) {
	a, _ := c.Get("123456")
	c.SuccessExit(a)
}

// @GET /TestResult
func (e *Example) TestResult(c *kapi.Context) (*MyResult, error) {
	return nil, nil
}

```



### 请求的参数可以使用类似继承的方式来声明参数

```go
type PageSize struct {
    Page int `query:"page,default=1"`
    Size int `query:"size,default=15"`
}

func (p *PageSize) GetLimit() (int, int) {
    return p.Size, (p.Page - 1) * p.Size
}
type GetBannerListReq struct {
	PageSize
}
```
### 可同时注册成多个HTTPMETHOD

```go
// @POST /testResult
// @GET /TestResult
func (e *Example) TestResult(c *kapi.Context) (*MyResult, error) {
	return nil, nil
}
```

目前还不能设置成一样的路由，即下面这种写法是不行的

```go
// @POST /TestResult
// @GET /TestResult
func (e *Example) TestResult(c *kapi.Context) (*MyResult, error) {
	return nil, nil
}
```



## 服务注入

可以在RegisterRouter之前注入你自己的服务, 注意依赖的顺序，比如服务B中需要服务A，那就需要先注入服务A 

```go
k.Map(config.AppConfig)
s3Svc := storage.NewS3Service()
k.MapTo(s3Svc, (*storage.IStorage)(nil))

k.RegisterRouter(...)
```

然后就可以在Controller中自动注入, 这是在RegisterRouter时进行注入的

```go
type MainController struct {
	CC    *config.Config   `inject:""`
	Store storage.IStorage `inject:""`
}
```

如果使用 inject.Default() 将服务注入到全局，也可以在任意地方进行手动注入

```
AA := new(config.Config)
err := inject.Default().Provide(AA)
```

Controller Method中也可以

```go
type MyService interface {
	Test() string
}
type MyServiceImpl struct {
}

func (m *MyServiceImpl) Test() string {
	return "MyServiceImpl"
}

// GetList 获取列表
// @GET /list
func (e *Example) GetList(
	c *kapi.Context,
	req *MyReq,
	svc MyService,
) {
	fmt.Println(req.Page, req.Size)
	fmt.Println(svc.Test())
	fmt.Println(e.CustomData)
	c.Success()
}
```



## 自定义统一的响应结构

`kapi.Context` 内置了一些快速返回数据的方法， 例如 `c.ListExit(count, list) ` 会返回下面这样的json

```json
{
  "code":"SUCCESS"
  "count": count,
  "data":[{},{}]
}
```

你可以实现 `kapi.IResultBuilder`接口来自定义这里的样式，默认的code字段是string

```go
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
```

只要注册进去即可

```go
k.RegisterResultBuilder(new(MyResultBuilder))
```

## pprof

需要自行注册

```go
import "github.com/gin-contrib/pprof"

pprof.Register(k.GetEngine(), "/pprof")

```













## 部署
`k build`后 `./bin/版本/系统_架构/`目录下的文件即为全部, 如果是自行编译, 则需要同时拷贝swagger.json和gen.gob以及config.toml.

当前主分支, 在 `k build -g` 后会覆盖 `swagger.json` 和 `gen.gob` ,`config.toml`则仅当输出目录下不存在`config.toml`时才会拷贝


## TODOList

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
- [x] 新的参数默认值，废弃旧的 default tag，改为使用gin的 `query:"name,default=hello"`
- [x] 增加一个注解用于注释返回结构类型, 特指只有一个Context的方法 @RESP
- [x] 增加multipart.FileHeader的支持
- [ ] 部分功能提取为单独包
- [ ] k编译打包时增加进度显示, 优化打包速度
- [ ] 精简引用包，减小体积
- [ ] 调整配置的读取
- [ ] 异常日志需要打印行数
- [ ] 注解加入可空默认的设定
- [ ] 加入markdown形式的文档
- [ ] 拦截器实现多个顺序执行机制（栈）
- [ ] 加入枚举支持

## 感谢

`https://github.com/xxjwxc/ginrpc` 大部分代码参考该项目, 例如代码解析\文档生成, 由于我需要一个方便单纯写写接口, 快速生成文档, 该项目无法满足,
而且也很难在基础上进行PR(改动较大, 并可能无法适应比较大众化的框架需求), 才有了魔改一顿的想法

# k
#### 介绍
这是本框架提供的命令行工具, 代码基本来自 `github.com/gogf/gf-cli`, 目前包含 安装、运行、编译三个部分， 后续会加入其它功能.

使用kapi进行开发, 建议同时使用k-cli, 由于kapi的swagger文档以及路由注册需要在开发环境运行后才会生成, 使用go自带的编译可能无法正常使用文档和注册路由

**目前移动到 https://github.com/linxlib/k** 仓库

