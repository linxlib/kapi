package kapi

import (
	"encoding/json"
	"fmt"
	"github.com/linxlib/kapi/binding"
	"github.com/linxlib/kapi/doc"
	ast_doc2 "github.com/linxlib/kapi/doc/ast_doc"
	"github.com/linxlib/kapi/doc/swagger"
	"github.com/linxlib/kapi/internal"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	binding2 "github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func buildRelativePath(prepath, routerPath string) string {
	if strings.HasSuffix(prepath, "/") {
		if strings.HasPrefix(routerPath, "/") {
			return prepath + strings.TrimPrefix(routerPath, "/")
		}
		return prepath + routerPath
	}

	if strings.HasPrefix(routerPath, "/") {
		return prepath + routerPath
	}

	return prepath + "/" + routerPath
}

func (b *KApi) handle(controller, method interface{}) gin.HandlerFunc {
	typ := reflect.TypeOf(method)
	hasReq := typ.NumIn() >= 2
	reqIsValue := true
	var req reflect.Value
	if hasReq {
		reqType := typ.In(1)
		req = reflect.New(reqType)
		if reqType.Kind() == reflect.Ptr {
			reqIsValue = false
			req = reflect.New(reqType.Elem())
		}
	}
	switch vt := method.(type) {
	case func(*Context):
		method = ContextInvoker(vt)
	}
	return func(context *gin.Context) {
		c := newContext(context)
		c.SetParent(b)
		c.Map(c)

		c.handler = method
		defer func() {
			if err := recover(); err != nil {
				b.option.recoverErrorFunc(err)
			}
		}()

		if i, ok := controller.(Interceptor); ok {
			i.Before(c)
		}
		if c.ctx.IsAborted() {
			return
		}
		if hasReq {
			if err := b.doBindReq(c, req.Interface()); err != nil {
				b.handleUnmarshalError(c, req, err)
				return
			}
			if reqIsValue {
				req = req.Elem()
			}
			c.Map(req.Interface())
		}
		returnValues, err := c.run()
		if err != nil {
			panic(fmt.Sprintf("unable to invoke the handler [%T]: %controller", method, err))
		}
		if c.ctx.IsAborted() {
			return
		}
		if len(returnValues) == 2 {
			resp := returnValues[0].Interface()
			rerr := returnValues[1].Interface().(error)
			if i, ok := controller.(Interceptor); ok {
				i.After(c)
			}
			if c.ctx.IsAborted() {
				return
			}
			if rerr != nil {
				c.ctx.PureJSON(GetResultFunc(RESULT_CODE_ERROR, rerr.Error(), 0, resp))
			} else {
				c.ctx.PureJSON(GetResultFunc(RESULT_CODE_SUCCESS, "", 0, resp))
			}
		}
	}
}

func (b *KApi) handleUnmarshalError(c *Context, req reflect.Value, err error) {
	var fields []string
	if _, ok := err.(validator.ValidationErrors); ok {
		for _, err := range err.(validator.ValidationErrors) {
			//TODO: 增加翻译选项
			tmp := fmt.Sprintf("%v:%v", internal.FindTag(req.Interface(), err.Field(), "json"), err.Tag())
			if len(err.Param()) > 0 {
				tmp += fmt.Sprintf("[%v](but[%v])", err.Param(), err.Value())
			}
			fields = append(fields, tmp)
		}
	} else if _, ok := err.(*json.UnmarshalTypeError); ok {
		err := err.(*json.UnmarshalTypeError)
		tmp := fmt.Sprintf("%v:%v(but[%v])", err.Field, err.Type.String(), err.Value)
		fields = append(fields, tmp)

	} else {
		fields = append(fields, err.Error())
	}

	c.ctx.PureJSON(GetResultFunc(RESULT_CODE_FAIL, fmt.Sprintf("req param : %v", strings.Join(fields, ";")), 0, nil))
	return
}

func (b *KApi) doBindReq(c *Context, v interface{}) error {
	if err := c.ctx.ShouldBindHeader(v); err != nil {
		if err != io.EOF {
			if _, ok := err.(validator.ValidationErrors); ok {

			} else {
				internal.Log.Errorln("ShouldBindHeader:", err)
				return err
			}
		}
	}

	if err := c.ctx.ShouldBindUri(v); err != nil {
		if err != io.EOF {
			if _, ok := err.(validator.ValidationErrors); ok {

			} else {
				internal.Log.Errorln("ShouldBindUri:", err)
				return err
			}

		}
	}
	if err := binding.Path.Bind(c.ctx, v); err != nil {
		if err != io.EOF {
			if _, ok := err.(validator.ValidationErrors); ok {

			} else {
				internal.Log.Errorln("Path.Bind:", err)
				return err
			}

		}
	}

	if err := c.ctx.ShouldBindWith(v, binding.Query); err != nil {
		if err != io.EOF {
			if _, ok := err.(validator.ValidationErrors); ok {

			} else {
				internal.Log.Errorln("ShouldBindWith.Query:", err)
				return err
			}

		}

	}
	if c.ctx.ContentType() == "multipart/form-data" {
		if err := c.ctx.ShouldBindWith(v, binding2.FormMultipart); err != nil {
			if err != io.EOF {
				if _, ok := err.(validator.ValidationErrors); ok {

				} else {
					internal.Log.Errorln("ShouldBindWith.FormMultipart:", err)
					return err
				}

			}

		}
	}

	if err := c.ctx.ShouldBindJSON(v); err != nil {
		if err != io.EOF {
			internal.Log.Errorln("body EOF:", err)
			return err
		}

	}
	return nil
}

func (b *KApi) analysisController(controller interface{}, model *doc.Model, modPkg string, modFile string) {
	controllerRefVal := reflect.ValueOf(controller)
	internal.Log.Debugf("解析 --> %s", controllerRefVal.Type().String())
	controllerType := reflect.Indirect(controllerRefVal).Type()
	controllerPkgPath := controllerType.PkgPath()
	controllerName := controllerType.Name()
	astDoc := ast_doc2.NewAstDoc(modPkg, modFile)
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
					routeInfo.AddFunc(controllerName+"/"+method.Name, mc.RouterPath, mc.Methods)
					if b.option.Server.NeedDoc {
						model.AddOne(controllerScheme.TagName, mc.RouterPath,
							mc.Methods, mc.Summary, mc.Description,
							siReq, siResp,
							controllerScheme.TokenHeader, mc.IsDeprecated)
					}

				}

			}
		}
	}
}

// analysisControllers gen out the Registered config info  by struct object,[prepath + objname.]
func (b *KApi) analysisControllers(router gin.IRoutes, controllers ...interface{}) bool {
	//TODO: groupPath也要加入到文档的路由中
	modPkg, modFile, isFind := ast_doc2.GetModuleInfo(2)
	if !isFind {
		return false
	}

	groupPath := b.BasePath(router)
	newDoc := doc.NewDoc(groupPath)
	for _, c := range controllers {
		b.analysisController(c, newDoc, modPkg, modFile)
	}

	if b.option.Server.NeedDoc {
		b.addDocModel(newDoc)
	}
	return true
}

func (b *KApi) addDocModel(model *doc.Model) {
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
		tag := swagger.Tag{Name: theTag}
		b.doc.AddTag(tag)
		//TODO: 重名方法 但是 METHOD不一样的情况
		for _, tagControllerMethod := range tagControllers {
			var p swagger.Param
			p.Tags = []string{theTag}
			p.Summary = tagControllerMethod.Summary
			p.Description = tagControllerMethod.Description

			myreqRef := ""
			p.Parameters = make([]swagger.Element, 0)
			p.Deprecated = tagControllerMethod.IsDeprecated

			if tagControllerMethod.TokenHeader != "" {
				p.Parameters = append(p.Parameters, swagger.Element{
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
					case doc.ParamTypeHeader:
						p.Parameters = append(p.Parameters, swagger.Element{
							In:          "header",
							Name:        item.Name,
							Description: item.Note,
							Required:    item.Required,
							Type:        swagger.GetKvType(item.Type, false, true),
							Schema:      nil,
							Default:     item.Default,
						})
					case doc.ParamTypeQuery:
						p.Parameters = append(p.Parameters, swagger.Element{
							In:          "query",
							Name:        item.Name,
							Description: item.Note,
							Required:    item.Required,
							Type:        swagger.GetKvType(item.Type, false, true),
							Schema:      nil,
							Default:     item.Default,
						})
					case doc.ParamTypeForm:
						t := swagger.GetKvType(item.Type, false, true)
						if item.IsFile {
							t = "file"
						}
						p.Parameters = append(p.Parameters, swagger.Element{
							In:          "formData",
							Name:        item.Name,
							Description: item.Note,
							Required:    item.Required,
							Type:        t,
							Schema:      nil,
							Default:     item.Default,
						})
					case doc.ParamTypePath:
						p.Parameters = append(p.Parameters, swagger.Element{
							In:          "path",
							Name:        item.Name,
							Description: item.Note,
							Required:    item.Required,
							Type:        swagger.GetKvType(item.Type, false, true),
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
							p.Parameters = append(p.Parameters, swagger.Element{
								In:          "body",
								Name:        tagControllerMethod.Req.Name,
								Description: item.Note,
								Required:    true,
								Schema: &swagger.Schema{
									Ref: myreqRef,
								},
							})
						}

					}

				}

			}

			if tagControllerMethod.Resp != nil {
				p.Responses = make(map[string]swagger.Resp)
				if len(tagControllerMethod.Resp.Items) > 0 {
					for _, item := range tagControllerMethod.Resp.Items {
						if tagControllerMethod.Resp.IsArray || item.IsArray {
							p.Responses["200"] = swagger.Resp{
								Description: "成功返回",
								Schema: map[string]interface{}{
									"type": "array",
									"items": map[string]string{
										"$ref": "#/definitions/" + tagControllerMethod.Resp.Name,
									},
								},
							}
						} else {
							p.Responses["200"] = swagger.Resp{
								Description: "成功返回",
								Schema: map[string]interface{}{
									"$ref": "#/definitions/" + tagControllerMethod.Resp.Name,
								},
							}
						}
					}
				} else {
					p.Responses["200"] = swagger.Resp{
						Description: "成功返回",
						Schema: map[string]interface{}{
							"type": tagControllerMethod.Resp.Name,
						},
					}
				}

			}

			url := buildRelativePath(model.Group, tagControllerMethod.RouterPath)
			for _, s := range tagControllerMethod.Methods {
				p.OperationID = s + "_" + strings.ReplaceAll(tagControllerMethod.RouterPath, "/", "_")
				b.doc.AddPatch2(url, p, s)
			}

		}
	}
}

func (b *KApi) BasePath(router gin.IRoutes) string {
	switch r := router.(type) {
	case *gin.RouterGroup:
		return r.BasePath()
	case *gin.Engine:
		return r.BasePath()
	}
	return ""
}

// register 注册路由到gin
func (b *KApi) register(router gin.IRoutes, cList ...interface{}) bool {
	if b.genFlag {
		return true
	}
	mp := routeInfo.getInfo()
	for _, c := range cList {
		refTyp := reflect.TypeOf(c)
		refVal := reflect.ValueOf(c)
		t := reflect.Indirect(refVal).Type()
		objName := t.Name()

		// Install the methods
		for m := 0; m < refTyp.NumMethod(); m++ {

			method := refTyp.Method(m)
			//_, _b := b.checkMethodParamCount(method.Type, true)
			if v, ok := mp[objName+"/"+method.Name]; ok {
				for _, v1 := range v {
					methods := strings.Join(v1.GenComment.Methods, ",")
					internal.Log.Debugf("%6s  %-20s --> %s", methods, v1.GenComment.RouterPath, t.PkgPath()+".(*"+objName+")."+method.Name)
					_ = b.registerMethodToRouter(router,
						v1.GenComment.Methods,
						v1.GenComment.RouterPath,
						refVal.Interface(),
						refVal.Method(m).Interface())
				}
			}
		}
	}
	return true
}

func (b *KApi) registerMethodToRouter(router gin.IRoutes, httpMethod []string, relativePath string, controller, method interface{}) error {
	call := b.handle(controller, method)
	for _, v := range httpMethod {
		switch strings.ToUpper(v) {
		case http.MethodPost:
			router.POST(relativePath, call)
		case http.MethodGet:
			router.GET(relativePath, call)
		case http.MethodDelete:
			router.DELETE(relativePath, call)
		case http.MethodPatch:
			router.PATCH(relativePath, call)
		case http.MethodPut:
			router.PUT(relativePath, call)
		case http.MethodOptions:
			router.OPTIONS(relativePath, call)
		case http.MethodHead:
			router.HEAD(relativePath, call)
		case "ANY":
			router.Any(relativePath, call)
		default:
			return fmt.Errorf("请求方式:[%controller] 不支持", httpMethod)
		}
	}

	return nil
}

// registerHandlerObj 注册路由方法
//func (b *KApi) registerHandlerObj(router gin.IRoutes, httpMethod []string, relativePath, methodName string, tvl, obj reflect.Value) error {
//	call := b.handlerFuncObj(tvl, obj, methodName)
//	for _, v := range httpMethod {
//		switch strings.ToUpper(v) {
//		case http.MethodPost:
//			router.POST(relativePath, call)
//		case http.MethodGet:
//			router.GET(relativePath, call)
//		case http.MethodDelete:
//			router.DELETE(relativePath, call)
//		case http.MethodPatch:
//			router.PATCH(relativePath, call)
//		case http.MethodPut:
//			router.PUT(relativePath, call)
//		case http.MethodOptions:
//			router.OPTIONS(relativePath, call)
//		case http.MethodHead:
//			router.HEAD(relativePath, call)
//		case "ANY":
//			router.Any(relativePath, call)
//		default:
//			return fmt.Errorf("请求方式:[%v] 不支持", httpMethod)
//		}
//	}
//
//	return nil
//}
