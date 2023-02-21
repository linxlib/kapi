# KAPI

## description
base on gin, support comments like //@METHOD, generate swagger doc automatically

the final target: make you write http api server quickly. 

PS: not stable now, use it in caution

## feature
1. mark and register route with comments
2. write codes just like Asp .Net Core's Controller
3. Dependency Inject using `https://github.com/linxlib/inject`
4. less comments to generate swagger doc
5. a built-in modified swagger ui(api path can be copied one-click)


## golang version 
golang > 1.18


## start



## comments support

use it like //@XXX parameter... 

| comment |location|description|  default  | optional | comment optional |
|:--:|:----:| :----: |:----:|:--:|----|
| @TAG  | Struct | the tag of swagger |注释|  | √ |
| @AUTH | Struct | HTTP Header name for authorization  |Authorization| √ | √ |
| @ROUTE | Struct | route prefix of api |None|  | √ |
| @HTTPMETHOD | Method | http method ||  |  |
| @RESP | Method | specify the model of result body ||  |  |



## sample

```go
// @GET /test
func (e *Example) TestPure(c *kapi.Context) {
	a, _ := c.Get("123456")
	c.SuccessExit(a)
}
```

can be

```go
// @GET /test
func (e *Example) TestPure(c *kapi.Context,req *MyRequest) {
	a, _ := c.Get("123456")
	c.SuccessExit(a)
}
```

kapi will analysis MyRequest，and show it in swagger

```go
type MyRequest struct {
  Page int `query:"page,default=1"` // you can set a default value for a parameter. only take effect in swagger
	Size int `query:"size,default=15"` // this parameter will be bind from query
  MyHeader string `header:"MyHeader"`
  MyBodyParam1 string `json:"myBodyParam1"`
  MyBodyParam2 int `json:"myBodyParam2"`
  MyPathParam int `path:"id" binding:"required"` //this parameter is required
  MyFormParam string `form:"myForm"` 
}
```

also

```go
type Form struct {
	FileHeader multipart.FileHeader `form:"file"` //you will see a file upload component in swagger ui
}
```

use `@RESP` mark response body's Model， the second way of writing belows are also ok

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



### request model can inherit from another struct

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
### register multiple HTTPMETHOD on the same time

the route should be different
```go
// @POST /testResult
// @GET /TestResult
func (e *Example) TestResult(c *kapi.Context) (*MyResult, error) {
	return nil, nil
}
```

belows are not suggested

```go
// @POST /TestResult
// @GET /TestResult
func (e *Example) TestResult(c *kapi.Context) (*MyResult, error) {
	return nil, nil
}
```



## Dependency Inject

register your service before `RegisterRouter`, you should handle dependency order by yourself

```go
k.Map(config.AppConfig)
s3Svc := storage.NewS3Service()
k.MapTo(s3Svc, (*storage.IStorage)(nil))

k.RegisterRouter(...)
```

use it in Controller

```go
type MainController struct {
	CC    *config.Config   `inject:""`
	Store storage.IStorage `inject:""`
}
```

you can also use `inject.Default()` to register a service globally

```
AA := new(config.Config)
err := inject.Default().Provide(AA)
```

Controller Method can supported DI

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
	svc MyService, //here
) {
	fmt.Println(req.Page, req.Size)
	fmt.Println(svc.Test())
	fmt.Println(e.CustomData)
	c.Success()
}
```



## customize the result model

`kapi.Context` has some method for easy return data， eg. `c.ListExit(count, list) ` will return like 

```json
{
  "code":"SUCCESS",
  "count": count,
  "data":[{},{}]
}
```

implement `kapi.IResultBuilder` to customize it.  

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

just register it

```go
k.RegisterResultBuilder(new(MyResultBuilder))
```

## pprof

```go
import "github.com/gin-contrib/pprof"

pprof.Register(k.GetEngine(), "/pprof")

```

## deploy

The "k" cli tool provides release feature, see "k" section below.


## TODOList
see README_CN.md for details.

## Thanks

the initial version is from `https://github.com/xxjwxc/ginrpc`


# k
### Description
a cli tool for KAPI.

### Features
1. initialize project struct
2. run project and restart project on file changed
3. build project
4. generate some codes

### Installation

```shell
go install github.com/linxlib/kapi/cmd/k@latest
```

### Usage

```bash
k -h 
k add -h
k build -h
k init -h
k run -h
```

![JetBrains Logo (Main) logo](https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png)