package kapi

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/ast_doc"
	binding3 "github.com/linxlib/kapi/internal/binding"
	doc2 "github.com/linxlib/kapi/internal/doc"
	swagger2 "github.com/linxlib/kapi/internal/swagger"
	"io"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	binding2 "github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

//Interceptor implement this to intercept controller method
type Interceptor interface {
	Before(*Context)
	After(*Context)
}

type HeaderAuth interface {
	HeaderAuth(c *Context)
}

type BeforeBind interface {
	BeforeBind(c *Context)
}

type AfterBind interface {
	AfterBind(c *Context)
}

type BeforeCall interface {
	BeforeCall(c *Context)
}

type AfterCall interface {
	AfterCall(c *Context)
}

type OnPanic interface {
	OnPanic(c *Context, err interface{})
}

type OnError interface {
	OnError(c *Context, err error)
}

type OnValidationError interface {
	OnValidationError(c *Context, err error)
}

type OnUnmarshalError interface {
	OnUnmarshalError(c *Context, err error)
}

// ContextInvoker method like this will be FastInvoke by inject package(not use reflect)
type ContextInvoker func(ctx *Context)

func (invoke ContextInvoker) Invoke(params []interface{}) ([]reflect.Value, error) {
	invoke(params[0].(*Context))
	return nil, nil
}

//handle get gin.HandlerFunc of a controller method
func (b *KApi) handle(controller, method interface{}) gin.HandlerFunc {
	typ := reflect.TypeOf(method)
	//TODO:
	hasReq := typ.NumIn() >= 2
	reqIsValue := true

	switch vt := method.(type) {
	case func(*Context):
		method = ContextInvoker(vt)
	}
	return func(context *gin.Context) {
		c := newContext(context)
		c.inj.SetParent(b)
		c.Map(c) //inject Context
		defer func() {
			if err := recover(); err != nil {
				b.option.recoverErrorFunc(err)
				if i, ok := controller.(OnPanic); ok {
					i.OnPanic(c, err)
				}
			}
		}()

		if i, ok := controller.(HeaderAuth); ok {
			i.HeaderAuth(c)
		}
		if c.IsAborted() {
			return
		}
		if hasReq {
			var req reflect.Value
			reqType := typ.In(1)
			req = reflect.New(reqType)
			if reqType.Kind() == reflect.Ptr {
				reqIsValue = false
				req = reflect.New(reqType.Elem())
			}
			if i, ok := controller.(BeforeBind); ok {
				i.BeforeBind(c)
			}
			if err := b.doBindReq(c, req.Interface()); err != nil {
				b.handleUnmarshalError(c, err)
				return
			}
			if reqIsValue {
				req = req.Elem()
			}
			c.Map(req.Interface())
			if i, ok := controller.(AfterBind); ok {
				i.AfterBind(c)
			}
		}
		if i, ok := controller.(BeforeCall); ok {
			i.BeforeCall(c)
		}
		returnValues, err := c.inj.Invoke(method)
		if err != nil {
			panic(fmt.Sprintf("unable to invoke the handler [%T]: %controller", method, err))
		}
		if c.IsAborted() {
			return
		}
		if i, ok := controller.(AfterCall); ok {
			i.AfterCall(c)
		}
		if c.IsAborted() {
			return
		}
		if len(returnValues) == 2 {
			resp := returnValues[0].Interface()
			rerr := returnValues[1].Interface().(error)

			if rerr != nil {
				c.PureJSON(GetResultFunc(RESULT_CODE_ERROR, rerr.Error(), 0, resp))
			} else {
				c.PureJSON(GetResultFunc(RESULT_CODE_SUCCESS, "", 0, resp))
			}
		}
	}
}

func (b *KApi) handleUnmarshalError(c *Context, err error) {
	var fields []string
	if v, ok := err.(validator.ValidationErrors); ok {
		for _, err := range v {
			tmp := err.Translate(trans)
			fields = append(fields, tmp)
		}
	} else if _, ok := err.(*json.UnmarshalTypeError); ok {
		err := err.(*json.UnmarshalTypeError)
		tmp := fmt.Sprintf("%v:%v(but[%v])", err.Field, err.Type.String(), err.Value)
		fields = append(fields, tmp)
	} else {
		fields = append(fields, err.Error())
	}
	c.PureJSON(GetResultFunc(RESULT_CODE_FAIL, strings.Join(fields, ";"), 0, nil))
	return
}

//doBindReq bind request to multi tag field
//  @param c
//  @param v request param
//
//  @return error
func (b *KApi) doBindReq(c *Context, v interface{}) error {
	if err := c.ShouldBindHeader(v); err != nil {
		if err != io.EOF {
			if _, ok := err.(validator.ValidationErrors); ok {

			} else {
				internal.Log.Errorln("ShouldBindHeader:", err)
				return err
			}
		}
	}

	if err := c.ShouldBindUri(v); err != nil {
		if err != io.EOF {
			if _, ok := err.(validator.ValidationErrors); ok {

			} else {
				internal.Log.Errorln("ShouldBindUri:", err)
				return err
			}

		}
	}
	if err := binding3.Path.Bind(c.Context, v); err != nil {
		if err != io.EOF {
			if _, ok := err.(validator.ValidationErrors); ok {

			} else {
				internal.Log.Errorln("Path.Bind:", err)
				return err
			}

		}
	}

	if err := c.ShouldBindWith(v, binding3.Query); err != nil {
		if err != io.EOF {
			if _, ok := err.(validator.ValidationErrors); ok {

			} else {
				internal.Log.Errorln("ShouldBindWith.Query:", err)
				return err
			}

		}

	}
	if c.ContentType() == "multipart/form-data" {
		if err := c.ShouldBindWith(v, binding2.FormMultipart); err != nil {
			if err != io.EOF {
				if _, ok := err.(validator.ValidationErrors); ok {

				} else {
					internal.Log.Errorln("ShouldBindWith.FormMultipart:", err)
					return err
				}

			}

		}
	}

	if err := c.ShouldBindJSON(v); err != nil {
		if err != io.EOF {
			internal.Log.Errorln("body EOF:", err)
			return err
		}

	}
	return nil
}

func (b *KApi) analysisController(controller interface{}, model *doc2.Model, modPkg string, modFile string) {
	controllerRefVal := reflect.ValueOf(controller)
	internal.Log.Debugf("%6s %s", ">", controllerRefVal.Type().String())
	controllerType := reflect.Indirect(controllerRefVal).Type()
	controllerPkgPath := controllerType.PkgPath()
	controllerName := controllerType.Name()
	astDoc := ast_doc.NewAstDoc(modPkg, modFile)
	if astDoc.FillPackage(controllerPkgPath) == nil {
		controllerScheme := astDoc.ResolveController(controllerName)
		refTyp := reflect.TypeOf(controller)
		// 遍历controller方法
		for m := 0; m < refTyp.NumMethod(); m++ {
			method := refTyp.Method(m)
			//_, _b := b.checkMethodParamCount(method.Type, true)
			if method.IsExported() {
				mc, siReq, siResp := astDoc.ResolveMethod(method.Name)
				if mc != nil {
					for k, v := range mc.Routes {
						routeInfo.AddFunc(controllerName+"/"+method.Name, k, v)
						if b.option.Server.NeedDoc {
							model.AddOne(controllerScheme.TagName, k,
								v, mc.Summary, mc.Description,
								siReq, siResp,
								controllerScheme.TokenHeader, mc.IsDeprecated)
						}
					}

				}

			}
		}
	}
}

// analysisControllers
func (b *KApi) analysisControllers(controllers ...interface{}) bool {
	start := time.Now()
	internal.Log.Debugf("analysis controllers...")
	//TODO: groupPath也要加入到文档的路由中
	modPkg, modFile, isFind := ast_doc.GetModuleInfo(2)
	if !isFind {
		return false
	}

	groupPath := b.engine.BasePath()
	newDoc := doc2.NewDoc(groupPath)
	for _, c := range controllers {
		b.analysisController(c, newDoc, modPkg, modFile)
	}

	if b.option.Server.NeedDoc {
		b.addDocModel(newDoc)
	}
	internal.Log.Debugf("elapsed time:%s", time.Now().Sub(start).String())
	return true
}

func (b *KApi) addDocModel(model *doc2.Model) {
	var tags []string
	for k, v := range model.TagControllers {
		for _, v1 := range v {
			b.doc.SetDefinition(model, v1.Req)
			b.doc.SetDefinition(model, v1.Resp)
		}
		tags = append(tags, k)
	}
	sort.Strings(tags)

	for _, theTag := range tags {
		tagControllers := model.TagControllers[theTag]
		tag := swagger2.Tag{Name: theTag}
		b.doc.AddTag(tag)
		//TODO: 重名方法 但是 METHOD不一样的情况
		for _, tagControllerMethod := range tagControllers {
			var p swagger2.Param
			p.Tags = []string{theTag}
			p.Summary = tagControllerMethod.Summary
			p.Description = tagControllerMethod.Description

			myreqRef := ""
			p.Parameters = make([]swagger2.Element, 0)
			p.Deprecated = tagControllerMethod.IsDeprecated

			if tagControllerMethod.TokenHeader != "" {
				p.Parameters = append(p.Parameters, swagger2.Element{
					In:          "header",
					Name:        tagControllerMethod.TokenHeader,
					Description: tagControllerMethod.TokenHeader,
					Required:    true,
					Type:        "string",
					Schema:      nil,
					Default:     "",
				})
			}

			if tagControllerMethod.Req != nil {
				for _, item := range tagControllerMethod.Req.Items {
					switch item.ParamType {
					case doc2.ParamTypeHeader:
						p.Parameters = append(p.Parameters, swagger2.Element{
							In:          "header",
							Name:        item.Name,
							Description: item.Note,
							Required:    item.Required,
							Type:        swagger2.GetKvType(item.Type, false, true),
							Schema:      nil,
							Default:     item.Default,
						})
					case doc2.ParamTypeQuery:
						p.Parameters = append(p.Parameters, swagger2.Element{
							In:          "query",
							Name:        item.Name,
							Description: item.Note,
							Required:    item.Required,
							Type:        swagger2.GetKvType(item.Type, false, true),
							Schema:      nil,
							Default:     item.Default,
						})
					case doc2.ParamTypeForm:
						t := swagger2.GetKvType(item.Type, false, true)
						if item.IsFile {
							t = "file"
						}
						p.Parameters = append(p.Parameters, swagger2.Element{
							In:          "formData",
							Name:        item.Name,
							Description: item.Note,
							Required:    item.Required,
							Type:        t,
							Schema:      nil,
							Default:     item.Default,
						})
					case doc2.ParamTypePath:
						p.Parameters = append(p.Parameters, swagger2.Element{
							In:          "path",
							Name:        item.Name,
							Description: item.Note,
							Required:    item.Required,
							Type:        swagger2.GetKvType(item.Type, false, true),
							Schema:      nil,
							Default:     item.Default,
						})
					default:
						myreqRef = "#/definitions/" + tagControllerMethod.Req.Name
						exist := false
						for _, parameter := range p.Parameters {
							if parameter.Name == tagControllerMethod.Req.Name {
								exist = true
							}
						}
						if !exist {
							p.Parameters = append(p.Parameters, swagger2.Element{
								In:          "body",
								Name:        tagControllerMethod.Req.Name,
								Description: item.Note,
								Required:    true,
								Schema: &swagger2.Schema{
									Ref: myreqRef,
								},
							})
						}

					}

				}

			}

			if tagControllerMethod.Resp != nil {
				p.Responses = make(map[string]swagger2.Resp)
				if len(tagControllerMethod.Resp.Items) > 0 {
					for _, item := range tagControllerMethod.Resp.Items {
						if tagControllerMethod.Resp.IsArray || item.IsArray {
							p.Responses["200"] = swagger2.Resp{
								Description: "成功返回",
								Schema: map[string]interface{}{
									"type": "array",
									"items": map[string]string{
										"$ref": "#/definitions/" + tagControllerMethod.Resp.Name,
									},
								},
							}
						} else {
							p.Responses["200"] = swagger2.Resp{
								Description: "成功返回",
								Schema: map[string]interface{}{
									"$ref": "#/definitions/" + tagControllerMethod.Resp.Name,
								},
							}
						}
					}
				} else {
					p.Responses["200"] = swagger2.Resp{
						Description: "成功返回",
						Schema: map[string]interface{}{
							"type": tagControllerMethod.Resp.Name,
						},
					}
				}

			}

			url := internal.BuildRelativePath(model.Group, tagControllerMethod.RouterPath)
			p.OperationID = tagControllerMethod.Method + "_" + strings.ReplaceAll(tagControllerMethod.RouterPath, "/", "_")
			b.doc.AddPatch2(url, p, tagControllerMethod.Method)

		}
	}
}

// register 注册路由到gin
func (b *KApi) register(router *gin.Engine, cList ...interface{}) {
	if b.genFlag {
		return
	}
	start := time.Now()
	internal.Log.Debug("register controllers..")
	mp := routeInfo.getInfo()
	for _, c := range cList {
		refTyp := reflect.TypeOf(c)
		refVal := reflect.ValueOf(c)
		t := reflect.Indirect(refVal).Type()
		objName := t.Name()

		// Install the Method
		for m := 0; m < refTyp.NumMethod(); m++ {
			method := refTyp.Method(m)
			//_, _b := b.checkMethodParamCount(method.Type, true)
			if v, ok := mp[objName+"/"+method.Name]; ok {
				for _, v1 := range v {
					internal.Log.Debugf("%6s  %-20s --> %s", v1.Method, v1.RouterPath, t.PkgPath()+">"+objName+">"+method.Name)
					err := b.registerMethodToRouter(router,
						v1.Method,
						v1.RouterPath,
						refVal.Interface(),
						refVal.Method(m).Interface())
					if err != nil {
						internal.Log.Errorln(err)
					}
				}
			}
		}
	}
	internal.Log.Debugf("elapsed time:%s", time.Now().Sub(start).String())
}

//registerMethodToRouter register to gin router
//  @param router
//  @param httpMethod
//  @param relativePath route path
//  @param controller
//  @param method
//
//  @return error
func (b *KApi) registerMethodToRouter(router *gin.Engine, httpMethod string, relativePath string, controller, method interface{}) error {
	call := b.handle(controller, method)

	switch strings.ToUpper(httpMethod) {
	case "POST":
		router.POST(relativePath, call)
	case "GET":
		router.GET(relativePath, call)
	case "DELETE":
		router.DELETE(relativePath, call)
	case "PATCH":
		router.PATCH(relativePath, call)
	case "PUT":
		router.PUT(relativePath, call)
	case "OPTIONS":
		router.OPTIONS(relativePath, call)
	case "HEAD":
		router.HEAD(relativePath, call)
	case "ANY":
		router.Any(relativePath, call)
	default:
		return fmt.Errorf("http method:[%v --> %s] 不支持", httpMethod, relativePath)
	}

	return nil
}

// genRouterCode 生成gen.gob
func (b *KApi) genRouterCode() {
	if !b.option.Server.Debug || b.doc == nil {
		return
	}
	internal.Log.Infoln("write out gen.gob")
	routeInfo.SetApiBody(*b.doc.Client)
	routeInfo.writeOut()
}

var (
	uni   *ut.UniversalTranslator
	trans ut.Translator
)

func init() {
	chinese := zh.New()
	uni = ut.New(chinese, chinese)
	trans, _ = uni.GetTranslator("zh")
	validate := binding3.Validator.Engine().(*validator.Validate)
	_ = zh_translations.RegisterDefaultTranslations(validate, trans)
}
