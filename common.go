package kapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-openapi/spec"
	binding3 "github.com/linxlib/binding"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/ast_parser"
	"github.com/linxlib/kapi/internal/comment_parser"
	"github.com/linxlib/kapi/internal/openapi"
	"io"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	binding2 "github.com/gin-gonic/gin/binding"
)

// Interceptor implement this to intercept controller method
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

// handle get gin.HandlerFunc of a controller method
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
		c := newContext(context, b)
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
				c.PureJSON(c.ResultBuilder.OnErrorDetail(rerr.Error(), resp))
			} else {
				c.PureJSON(c.ResultBuilder.OnData("", 0, resp))
			}
		}
	}
}

func (b *KApi) handleUnmarshalError(c *Context, err error) {
	var fields []string
	if v, ok := binding3.HandleValidationErrors(err); ok {
		fields = v
	} else if _, ok := err.(*json.UnmarshalTypeError); ok {
		err := err.(*json.UnmarshalTypeError)
		tmp := fmt.Sprintf("%v:%v(but[%v])", err.Field, err.Type.String(), err.Value)
		fields = append(fields, tmp)
	} else {
		fields = append(fields, err.Error())
	}
	c.PureJSON(c.ResultBuilder.OnFail(strings.Join(fields, ";"), fields))
	return
}

// doBindReq bind request to multi tag field
//
//	@param c
//	@param v request param
//
//	@return error
func (b *KApi) doBindReq(c *Context, v interface{}) error {
	if err := c.ShouldBindHeader(v); err != nil {
		if err != io.EOF {
			if _, ok := binding3.HandleValidationErrors(err); ok {

			} else {
				Errorf("ShouldBindHeader:%s", err)
				return err
			}
		}
	}

	if err := c.ShouldBindUri(v); err != nil {
		if err != io.EOF {
			if _, ok := binding3.HandleValidationErrors(err); ok {

			} else {
				Errorf("ShouldBindUri: %s", err)
				return err
			}

		}
	}
	if err := binding3.Path.Bind(c.Context, v); err != nil {
		if err != io.EOF {
			if _, ok := binding3.HandleValidationErrors(err); ok {

			} else {
				Errorf("Path.Bind:%s", err)
				return err
			}

		}
	}

	if err := c.ShouldBindWith(v, binding3.Query); err != nil {
		if err != io.EOF {
			if _, ok := binding3.HandleValidationErrors(err); ok {

			} else {
				Errorf("ShouldBindWith.Query:%s", err)
				return err
			}

		}

	}
	if c.ContentType() == "multipart/form-data" {
		if err := c.ShouldBindWith(v, binding2.FormMultipart); err != nil {
			if err != io.EOF {
				if _, ok := binding3.HandleValidationErrors(err); ok {

				} else {
					Errorf("ShouldBindWith.FormMultipart:%s", err)
					return err
				}

			}

		}
	}

	if err := c.ShouldBindJSON(v); err != nil {
		if err != io.EOF {
			Errorf("body EOF:%s", err)
			return err
		}

	}
	return nil
}

func (b *KApi) analysisController2(controller interface{}, modPkg string, modFile string) {

	controllerRefVal := reflect.ValueOf(controller)
	Debugf("%6s %s", ">", controllerRefVal.Type().String())
	controllerType := reflect.Indirect(controllerRefVal).Type()
	controllerPkgPath := controllerType.PkgPath()
	parser := ast_parser.NewParser(modPkg, modFile)
	f, err := parser.Parse(controllerPkgPath, controllerType.Name())
	if err != nil {
		Errorf("%+v", err)
		return
	}
	controllerStruct := f.Structs[0]
	controllerParser := comment_parser.NewParser(controllerStruct.Name, controllerStruct.Docs)
	cp := controllerParser.Parse("")
	for _, method := range controllerStruct.Methods {
		if !method.Private {
			p := comment_parser.NewParser(method.Name, method.Docs)
			methodComment := p.Parse(cp.Route)
			for m, r := range methodComment.Routes {
				b.routeInfo.AddFunc(controllerType.Name()+"/"+method.Name, m, r)
				if b.option.Server.NeedDoc {
					if cp.Deprecated {
						methodComment.Deprecated = true //deprecate all method
					}
					var tag = cp.Summary
					if cp.Tag == "" {
						tag = cp.Summary
					}
					tagBuilder := openapi.NewTagBuilder()
					tagBuilder.Name = tag
					b.doc.AddTag(tagBuilder.Build())
					var requestParams = make([]*openapi.ParameterBuilder, 0)
					var responseParams = make([]*openapi.ResponseBuilder, 0)
					var sReq *ast_parser.Struct
					// has @REQ but request param not defined
					if methodComment.HasReq && len(methodComment.RequestType) > 0 && len(method.Params) <= 1 {
						pkg := ""
						if len(methodComment.RequestType) > 0 {
							pkg = methodComment.RequestType[0]
						}
						t := ""
						if len(methodComment.RequestType) > 1 {
							t = methodComment.RequestType[1]
						}
						f1, _ := parser.Parse(pkg, t)
						sReq = f1.Structs[0]
					} else if len(method.Params) > 1 {
						sReq = method.Params[1].Struct
					}
					if sReq != nil {
						// build json body scheme and add it into definitions
						fds := sReq.GetAllFieldsByTag("json")

						bodyParameter := openapi.NewParameterBuilder()
						bodyParameter.WithDescription(strings.Join(sReq.Docs, "\n")).
							WithLocation("body").
							AsRequired().Named(sReq.Name)
						bodyParameter.Ref = openapi.NewRefBuilder("#/definitions/" + sReq.Name).Build()

						requestParams = append(requestParams, bodyParameter)

						bodyDefineSchema := openapi.NewSchemaBuilder()
						bodyDefineSchema.WithDescription(strings.Join(sReq.Docs, "\n"))
						bodyDefineSchema.Typed("object", "")
						for _, field := range fds {
							bodyFieldSchema := openapi.NewSchemaBuilder()
							bodyFieldSchema.WithDescription(field.Comment)
							if field.GetTag("default") != "" {
								bodyFieldSchema.WithDefault(field.GetTag("default"))
							}
							if strings.HasPrefix(field.Type, "[]") {
								bodyFieldSchema.Typed("array", "")
								ss := openapi.NewSchemaBuilder()
								ss.AddType(internal.GetType(strings.TrimPrefix(field.Type, "[]")), internal.GetFormat(field.Type, ""))
								ss.Build()
								bodyFieldSchema.Items = &spec.SchemaOrArray{}
								bodyFieldSchema.Items.Schema = &ss.Schema
							} else {
								if field.GetTag("default") != "" {
									bodyFieldSchema.WithDefault(field.GetTag("default"))
								}
								bodyFieldSchema.Typed(internal.GetType(field.Type), internal.GetFormat(field.Type, ""))
							}

							//bodyFieldSchema.WithExample()
							if strings.Contains(field.GetTag("v"), "required") {
								bodyDefineSchema.AddRequired(field.Name)
							}
							if field.GetTag("json") == "" {
								bodyDefineSchema.SetProperty(field.Name, bodyFieldSchema.Build())
							} else {
								bodyDefineSchema.SetProperty(field.GetTag("json"), bodyFieldSchema.Build())
							}

						}
						b.doc.AddDefinitions(sReq.Name, bodyDefineSchema.Build())

						queryFields := sReq.GetAllFieldsByTag("query")
						for _, field := range queryFields {
							queryParameter := openapi.NewParameterBuilder()
							queryParameter.WithLocation("query")
							queryParameter.Named(field.GetTag("query"))
							queryParameter.WithDescription(field.Comment)
							if field.GetTag("v") == "required" {
								queryParameter.AsRequired()
							}
							//TODO: get param type
							if strings.HasPrefix(field.Type, "[]") {
								queryParameter.Typed("array", "")
								queryParameter.Items = &spec.Items{}
								queryParameter.Items.Type = internal.GetType(strings.TrimPrefix(field.Type, "[]"))
							} else {
								if field.GetTag("default") != "" {
									queryParameter.WithDefault(field.GetTag("default"))
								}
								queryParameter.Typed(internal.GetType(field.Type), internal.GetFormat(field.Type, ""))
							}

							queryParameter.Build()
							requestParams = append(requestParams, queryParameter)
						}

						headerFields := sReq.GetAllFieldsByTag("header")
						for _, field := range headerFields {
							headerParameter := openapi.NewParameterBuilder()
							headerParameter.WithLocation("header")
							headerParameter.Named(field.GetTag("header"))
							headerParameter.WithDescription(field.Comment)
							if field.GetTag("v") == "required" {
								headerParameter.AsRequired()
							}
							//TODO: get param type
							if strings.HasPrefix(field.Type, "[]") {
								panic("slice header type is not supported")
								//headerParameter.Typed("array", "")
								//headerParameter.Items = &spec.Items{}
								//headerParameter.Items.Type = internal.GetType(strings.TrimPrefix(field.Type, "[]"))
							} else {
								if field.GetTag("default") != "" {
									headerParameter.WithDefault(field.GetTag("default"))
								}
								headerParameter.Typed(internal.GetType(field.Type), internal.GetFormat(field.Type, ""))
							}
							headerParameter.Build()
							requestParams = append(requestParams, headerParameter)
						}
						// TODO: 使用form时 POST请求会解析失败
						formFields := sReq.GetAllFieldsByTag("form")
						for _, field := range formFields {
							formParameter := openapi.NewParameterBuilder()
							formParameter.WithLocation("formData")
							formParameter.Named(field.GetTag("form"))
							formParameter.WithDescription(field.Comment)
							if field.GetTag("v") == "required" {
								formParameter.AsRequired()
							}
							//TODO: get param type
							if strings.HasPrefix(field.Type, "[]") {
								formParameter.Typed("array", "")
								formParameter.Items = &spec.Items{}
								formParameter.Items.Type = internal.GetType(strings.TrimPrefix(field.Type, "[]"))
							} else {
								if field.GetTag("default") != "" {
									formParameter.WithDefault(field.GetTag("default"))
								}
								formParameter.Typed(internal.GetType(field.Type), internal.GetFormat(field.Type, ""))
							}
							formParameter.Build()
							requestParams = append(requestParams, formParameter)
						}
						pathFields := sReq.GetAllFieldsByTag("path")
						for _, field := range pathFields {
							pathParameter := openapi.NewParameterBuilder()
							pathParameter.WithLocation("path")
							pathParameter.Named(field.GetTag("path"))
							pathParameter.WithDescription(field.Comment)
							if field.GetTag("v") == "required" {
								pathParameter.AsRequired()
							}
							//TODO: get param type
							if strings.HasPrefix(field.Type, "[]") {
								panic("slice path type is not supported")
								//pathParameter.Typed("array", "")
								//pathParameter.Items = &spec.Items{}
								//pathParameter.Items.Type = internal.GetType(strings.TrimPrefix(field.Type, "[]"))
							} else {
								if field.GetTag("default") != "" {
									pathParameter.WithDefault(field.GetTag("default"))
								}
								pathParameter.Typed(internal.GetType(field.Type), internal.GetFormat(field.Type, ""))
							}
							pathParameter.Build()
							requestParams = append(requestParams, pathParameter)
						}
					}

					var sResp *ast_parser.Struct

					if methodComment.HasResp && len(methodComment.ResultType) > 0 && len(method.Results) <= 0 {
						pkg := ""
						if len(methodComment.ResultType) > 0 {
							pkg = methodComment.ResultType[0]
						}
						t := ""
						if len(methodComment.ResultType) > 1 {
							t = methodComment.ResultType[1]
						}
						f2, _ := parser.Parse(pkg, t)
						sResp = f2.Structs[0]
					} else if len(method.Results) > 0 {
						sResp = method.Results[0].Struct
					}
					if sResp != nil {
						fds := sResp.GetAllFieldsByTag("json")

						response := openapi.NewResponseBuilder()
						response.WithDescription(strings.Join(sResp.Docs, "\n"))
						responseSchemaBuilder := openapi.NewSchemaBuilder()
						responseSchemaBuilder.Ref = openapi.NewRefBuilder("#/definitions/" + sResp.Name).Build()
						sch := responseSchemaBuilder.Build()
						response.WithSchema(&sch)
						response.Build()
						responseParams = append(responseParams, response)

						bodyDefineSchema := openapi.NewSchemaBuilder()
						bodyDefineSchema.WithDescription(strings.Join(sResp.Docs, "\n"))
						for _, field := range fds {
							bodyFieldSchema := openapi.NewSchemaBuilder()
							bodyFieldSchema.WithDescription(field.Comment)

							if strings.HasPrefix(field.Type, "[]") {
								bodyFieldSchema.Typed("array", "")
								ss := openapi.NewSchemaBuilder()
								ss.AddType(internal.GetType(strings.TrimPrefix(field.Type, "[]")), internal.GetFormat(field.Type, ""))
								ss.Build()
								bodyFieldSchema.Items = &spec.SchemaOrArray{}
								bodyFieldSchema.Items.Schema = &ss.Schema
							} else {
								if field.GetTag("default") != "" {
									bodyFieldSchema.WithDefault(field.GetTag("default"))
								}
								bodyFieldSchema.Typed(internal.GetType(field.Type), internal.GetFormat(field.Type, ""))
							}

							//bodyFieldSchema.WithExample()
							if strings.Contains(field.GetTag("v"), "required") {
								bodyDefineSchema.AddRequired(field.Name)
							}
							if field.GetTag("json") != "" {
								bodyDefineSchema.SetProperty(field.GetTag("json"), bodyFieldSchema.Build())
							} else {
								bodyDefineSchema.SetProperty(field.Name, bodyFieldSchema.Build())
							}

						}
						b.doc.AddDefinitions(sResp.Name, bodyDefineSchema.Build())
					}
					// 方法可能注册为多条路由
					for r, m := range methodComment.Routes {
						if strings.Contains(r, "{") || strings.Contains(r, "}") {
							panic("path route {path} not supported. use :path instead")
						}
						opBuilder := openapi.NewOperationBuilder()
						opBuilder.Deprecated = methodComment.Deprecated
						opBuilder.WithSummary(strings.Join(methodComment.Description, ","))
						opBuilder.WithID(method.Name)
						opBuilder.WithTags(tag)
						for _, param := range requestParams {
							if param.IsBuilt() {
								opBuilder.AddParam(param.Parameter)
							}
						}
						for _, param := range responseParams {
							if param.IsBuilt() {
								opBuilder.RespondsWith(200, param.Response)
							}
						}
						b.doc.AddRoute(m, r, opBuilder.Build())
					}
				}
			}

		}
	}
	//Infof("%#v", f)
	_ = f
}

//func (b *KApi) analysisController(controller interface{}, model *doc.Model, modPkg string, modFile string) {
//	controllerRefVal := reflect.ValueOf(controller)
//	Debugf("%6s %s", ">", controllerRefVal.Type().String())
//	controllerType := reflect.Indirect(controllerRefVal).Type()
//	controllerPkgPath := controllerType.PkgPath()
//	controllerName := controllerType.Name()
//	astDoc := ast_doc.NewAstDoc(modPkg, modFile)
//	if astDoc.FillPackage(controllerPkgPath) == nil {
//		controllerScheme := astDoc.ResolveController(controllerName)
//		refTyp := reflect.TypeOf(controller)
//		// 遍历controller方法
//		for m := 0; m < refTyp.NumMethod(); m++ {
//			method := refTyp.Method(m)
//			//_, _b := b.checkMethodParamCount(method.Type, true)
//			if method.IsExported() {
//				mc, siReq, siResp := astDoc.ResolveMethod(method.Name)
//				if mc != nil {
//					for k, v := range mc.Routes {
//						routeInfo.AddFunc(controllerName+"/"+method.Name, k, v)
//						if b.option.Server.NeedDoc {
//							model.AddOne(controllerScheme.TagName, k,
//								v, mc.Summary, mc.Description,
//								siReq, siResp,
//								controllerScheme.TokenHeader, mc.IsDeprecated)
//						}
//					}
//
//				}
//
//			}
//		}
//	}
//}

// GetModuleInfo 获取项目[module name] [根目录绝对地址]
func GetModuleInfo(n int) (string, string, bool) {
	index := n
	// 本包被引用时需要向上推2级查找main.go
	for {
		_, filename, _, ok := runtime.Caller(index)
		if ok {
			if strings.HasSuffix(filename, "runtime/asm_amd64.s") {
				index -= 2
				break
			}
			if strings.HasSuffix(filename, "runtime/asm_arm64.s") {
				index -= 2
				break
			}
			index++
		} else {
			panic(errors.New("package parsing failed:can not find main file"))
		}
	}

	_, filename, _, _ := runtime.Caller(index)
	filename = strings.Replace(filename, "\\", "/", -1) // change windows path delimiter '\' to unix path delimiter '/'
	for {
		n := strings.LastIndex(filename, "/")
		if n > 0 {
			filename = filename[0:n]
			if internal.FileIsExist(filename + "/go.mod") {
				return internal.GetMod(filename + "/go.mod"), filename, true
			}
		} else {
			break
			// panic(errors.New("package parsing failed:can not find module file[go.mod] , golang version must up 1.11"))
		}
	}

	// never reach
	return "", "", false
}

// analysisControllers
func (b *KApi) analysisControllers(controllers ...interface{}) bool {
	start := time.Now()
	Debugf("analysis controllers...")
	//TODO: groupPath也要加入到文档的路由中
	modPkg, modFile, isFind := GetModuleInfo(2)
	if !isFind {
		return false
	}

	//groupPath := b.engine.BasePath()
	for _, c := range controllers {
		b.analysisController2(c, modPkg, modFile)
		//b.analysisController(c, newDoc, modPkg, modFile)
	}

	//if b.option.Server.NeedDoc {
	//	b.addDocModel(newDoc)
	//}
	Debugf("elapsed time:%s", time.Now().Sub(start).String())
	return true
}

//func (b *KApi) addDocModel(model *doc.Model) {
//	var tags []string
//	for k, v := range model.TagControllers {
//		for _, v1 := range v {
//			b.doc.SetDefinition(model, v1.Req)
//			b.doc.SetDefinition(model, v1.Resp)
//		}
//		tags = append(tags, k)
//	}
//	sort.Strings(tags)
//
//	for _, theTag := range tags {
//		tagControllers := model.TagControllers[theTag]
//		tag := swagger2.Tag{Name: theTag}
//		b.doc.AddTag(tag)
//		//TODO: 重名方法 但是 METHOD不一样的情况
//		for _, tagControllerMethod := range tagControllers {
//			var p swagger2.Param
//			p.Tags = []string{theTag}
//			p.Summary = tagControllerMethod.Summary
//			p.Description = tagControllerMethod.Description
//
//			myreqRef := ""
//			p.Parameters = make([]swagger2.Element, 0)
//			p.Deprecated = tagControllerMethod.IsDeprecated
//
//			if tagControllerMethod.TokenHeader != "" {
//				p.Parameters = append(p.Parameters, swagger2.Element{
//					In:          "header",
//					Name:        tagControllerMethod.TokenHeader,
//					Description: tagControllerMethod.TokenHeader,
//					Required:    true,
//					Type:        "string",
//					Schema:      nil,
//					Default:     "",
//				})
//			}
//
//			if tagControllerMethod.Req != nil {
//				for _, item := range tagControllerMethod.Req.Items {
//					switch item.ParamType {
//					case doc.ParamTypeHeader:
//						p.Parameters = append(p.Parameters, swagger2.Element{
//							In:          "header",
//							Name:        item.Name,
//							Description: item.Note,
//							Required:    item.Required,
//							Type:        internal.GetKvType(item.Type, false, true),
//							Schema:      nil,
//							Default:     item.Default,
//						})
//					case doc.ParamTypeQuery:
//						p.Parameters = append(p.Parameters, swagger2.Element{
//							In:          "query",
//							Name:        item.Name,
//							Description: item.Note,
//							Required:    item.Required,
//							Type:        internal.GetKvType(item.Type, false, true),
//							Schema:      nil,
//							Default:     item.Default,
//						})
//					case doc.ParamTypeForm:
//						t := internal.GetKvType(item.Type, false, true)
//						if item.IsFile {
//							t = "file"
//						}
//						p.Parameters = append(p.Parameters, swagger2.Element{
//							In:          "formData",
//							Name:        item.Name,
//							Description: item.Note,
//							Required:    item.Required,
//							Type:        t,
//							Schema:      nil,
//							Default:     item.Default,
//						})
//					case doc.ParamTypePath:
//						p.Parameters = append(p.Parameters, swagger2.Element{
//							In:          "path",
//							Name:        item.Name,
//							Description: item.Note,
//							Required:    item.Required,
//							Type:        internal.GetKvType(item.Type, false, true),
//							Schema:      nil,
//							Default:     item.Default,
//						})
//					default:
//						myreqRef = "#/definitions/" + tagControllerMethod.Req.Name
//						exist := false
//						for _, parameter := range p.Parameters {
//							if parameter.Name == tagControllerMethod.Req.Name {
//								exist = true
//							}
//						}
//						if !exist {
//							p.Parameters = append(p.Parameters, swagger2.Element{
//								In:          "body",
//								Name:        tagControllerMethod.Req.Name,
//								Description: item.Note,
//								Required:    true,
//								Schema: &swagger2.Schema{
//									Ref: myreqRef,
//								},
//							})
//						}
//
//					}
//
//				}
//
//			}
//
//			if tagControllerMethod.Resp != nil {
//				p.Responses = make(map[string]swagger2.Resp)
//				if len(tagControllerMethod.Resp.Items) > 0 {
//					for _, item := range tagControllerMethod.Resp.Items {
//						if tagControllerMethod.Resp.IsArray || item.IsArray {
//							p.Responses["200"] = swagger2.Resp{
//								Description: "成功返回",
//								Schema: map[string]interface{}{
//									"type": "array",
//									"items": map[string]string{
//										"$ref": "#/definitions/" + tagControllerMethod.Resp.Name,
//									},
//								},
//							}
//						} else {
//							p.Responses["200"] = swagger2.Resp{
//								Description: "成功返回",
//								Schema: map[string]interface{}{
//									"$ref": "#/definitions/" + tagControllerMethod.Resp.Name,
//								},
//							}
//						}
//					}
//				} else {
//					p.Responses["200"] = swagger2.Resp{
//						Description: "成功返回",
//						Schema: map[string]interface{}{
//							"type": tagControllerMethod.Resp.Name,
//						},
//					}
//				}
//
//			}
//
//			url := buildRelativePath(model.Group, tagControllerMethod.RouterPath)
//			p.OperationID = tagControllerMethod.Method + "_" + strings.ReplaceAll(tagControllerMethod.RouterPath, "/", "_")
//			b.doc.AddPatch2(url, p, tagControllerMethod.Method)
//
//		}
//	}
//}

//func buildRelativePath(prepath, routerPath string) string {
//	if strings.HasSuffix(prepath, "/") {
//		if strings.HasPrefix(routerPath, "/") {
//			return prepath + strings.TrimPrefix(routerPath, "/")
//		}
//		return prepath + routerPath
//	}
//
//	if strings.HasPrefix(routerPath, "/") {
//		return prepath + routerPath
//	}
//
//	return prepath + "/" + routerPath
//}

// register 注册路由到gin
func (b *KApi) register(router *gin.Engine, cList ...interface{}) {
	start := time.Now()
	Debugf("register controllers..")
	mp := b.routeInfo.GetGenInfo().Routes
	for _, c := range cList {
		refTyp := reflect.TypeOf(c)
		refVal := reflect.ValueOf(c)
		t := reflect.Indirect(refVal).Type()
		objName := t.Name()
		b.Apply(c)

		// Install the Method
		for m := 0; m < refTyp.NumMethod(); m++ {
			method := refTyp.Method(m)
			//_, _b := b.checkMethodParamCount(method.Type, true)
			k := objName + "/" + method.Name
			for _, item := range mp {
				if item.Key == k {
					Debugf("%6s  %-30s --> %s", k, item.Method, t.PkgPath()+"."+objName+"."+method.Name)
					err := b.registerMethodToRouter(router,
						item.Method,
						item.RouterPath,
						refVal.Interface(),
						refVal.Method(m).Interface())
					if err != nil {
						Errorf("%s", err)
					}
				}
			}

		}
	}
	Debugf("elapsed time:%s", time.Now().Sub(start).String())
}

// registerMethodToRouter register to gin router
//
//	@param router
//	@param httpMethod
//	@param relativePath route path
//	@param controller
//	@param method
//
//	@return error
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
	Infof("write out gen.gob")
	b.routeInfo.SetApiBody(b.doc.Swagger)
	b.routeInfo.WriteOut()
}
