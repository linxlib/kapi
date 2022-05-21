package kapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/linxlib/kapi/ast"
	"github.com/linxlib/kapi/binding"
	"github.com/linxlib/kapi/doc"
	"github.com/linxlib/kapi/doc/swagger"
	"github.com/linxlib/kapi/internal"
	goast "go/ast"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// checkHandlerFunc 检查路由方法是否符合 返回参数个数
// isObj表示方法是否是struct下的方法 如果是则 typ.In(0) 是struct本身
func (b *KApi) checkHandlerFunc(typ reflect.Type, isObj bool) (int, bool) {
	offset := 0
	if isObj {
		offset = 1
	}
	paramCount := typ.NumIn() - offset
	if paramCount == 1 || paramCount == 2 { // 参数检查
		ctxType := typ.In(0 + offset)

		// go-gin 框架的默认方法
		if ctxType == reflect.TypeOf(&gin.Context{}) {
			return paramCount, true
		}

		// Customized context . 自定义的context
		if ctxType == b.customContextType {
			return paramCount, true
		}

		// maybe interface
		if b.customContextType.ConvertibleTo(ctxType) {
			return paramCount, true
		}

	}
	return paramCount, false
}

// handlerFuncObj Get and filter the parameters to be bound (object call type)
func (b *KApi) handlerFuncObj(tvl, obj reflect.Value, methodName string) gin.HandlerFunc { // 获取并过滤要绑定的参数(obj 对象类型)
	typ := tvl.Type()
	if typ.NumIn() == 2 { //1个参数的方法
		ctxType := typ.In(1)

		apiFun := func(c *gin.Context) interface{} { return c }
		if ctxType == b.customContextType { // Customized context . 自定义的context
			apiFun = b.apiFun
		} else if !(ctxType == reflect.TypeOf(&gin.Context{})) {
			internal.Log.Errorln("方法不受支持 " + runtime.FuncForPC(tvl.Pointer()).Name() + " !")
		}

		return func(c *gin.Context) {
			//加入recover处理
			defer func() {
				if err := recover(); err != nil {
					b.option.recoverErrorFunc(err)
				}
			}()

			bainfo, is := b.preCall(c, obj, nil, methodName)
			if !is {
				c.JSON(bainfo.RespCode, bainfo.Resp)
				return
			}

			var returnValues []reflect.Value
			returnValues = tvl.Call([]reflect.Value{obj, reflect.ValueOf(apiFun(c))})

			if returnValues != nil {
				bainfo.Resp = returnValues[0].Interface()
				rerr := returnValues[1].Interface()
				if rerr != nil {
					bainfo.Error = rerr.(error)
				}

				is = b.afterCall(bainfo, obj)
				if is {
					if bainfo.Error != nil {
						c.JSON(GetResultFunc(RESULT_CODE_ERROR, bainfo.Error.Error(), 0, bainfo.Resp))
					} else {
						c.JSON(GetResultFunc(RESULT_CODE_SUCCESS, "", 0, bainfo.Resp))
					}
				} else {
					if bainfo.Error != nil {
						c.JSON(GetResultFunc(RESULT_CODE_ERROR, bainfo.Error.Error(), 0, bainfo.Resp))
					} else {
						c.JSON(GetResultFunc(RESULT_CODE_ERROR, "", 0, bainfo.Resp))
					}
				}
			}

		}
	}

	// 自定义的context类型,带request 请求参数
	call, err := b.changeToGinHandlerFunc(tvl, obj, methodName)
	if err != nil { // Direct reporting error.
		panic(err)
	}

	return call
}

//preCall 调用前处理
func (b *KApi) preCall(c *gin.Context, obj reflect.Value, req interface{}, methodName string) (*InterceptorContext, bool) {
	info := &InterceptorContext{
		C:        &Context{c},
		FuncName: fmt.Sprintf("%v.%v", reflect.Indirect(obj).Type().Name(), methodName), // 函数名
		Req:      req,                                                                   // 调用前的请求参数
		Context:  context.Background(),                                                  // 占位参数，可用于存储其他参数，前后连接可用
	}
	//处理控制器的拦截器
	is := true

	if preObj, ok := obj.Interface().(Interceptor); ok { // 本类型
		is = preObj.Before(info)
	}
	//处理全局的拦截器
	if is && b.beforeAfter != nil {
		is = b.beforeAfter.Before(info)
	}
	return info, is
}

//afterCall 调用后处理
func (b *KApi) afterCall(info *InterceptorContext, obj reflect.Value) bool {
	is := true
	if bfobj, ok := obj.Interface().(Interceptor); ok { // 本类型
		is = bfobj.After(info)
	}
	if is && b.beforeAfter != nil {
		is = b.beforeAfter.After(info)
	}
	return is
}

// Custom context type with request parameters
func (b *KApi) changeToGinHandlerFunc(tvl, obj reflect.Value, methodName string) (func(*gin.Context), error) {
	typ := tvl.Type()

	if typ.NumOut() != 0 {
		if typ.NumOut() == 2 {
			if returnType := typ.Out(1); returnType != typeOfError {
				return nil, fmt.Errorf("方法 : %v , 第二返回值 %v 不是 error",
					runtime.FuncForPC(tvl.Pointer()).Name(), returnType.String())
			}
		} else {
			return nil, fmt.Errorf("方法 : %v 不受支持, 只支持两个返回值 (obj, error) 的方法", runtime.FuncForPC(tvl.Pointer()).Name())
		}
	}

	ctxType, reqType := typ.In(1), typ.In(2)

	reqIsGinCtx := false
	if ctxType == reflect.TypeOf(&gin.Context{}) {
		reqIsGinCtx = true
	}

	if !reqIsGinCtx && ctxType != b.customContextType && !b.customContextType.ConvertibleTo(ctxType) {
		return nil, errors.New("方法 " + runtime.FuncForPC(tvl.Pointer()).Name() + " 第一个参数不受支持!")
	}

	reqIsValue := true
	if reqType.Kind() == reflect.Ptr {
		reqIsValue = false
	}
	apiFun := func(c *gin.Context) interface{} { return c }
	if !reqIsGinCtx {
		apiFun = b.apiFun
	}

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				b.option.recoverErrorFunc(err)
			}
		}()

		req := reflect.New(reqType)
		if !reqIsValue {
			req = reflect.New(reqType.Elem())
		}
		if err := b.unmarshal(c, req.Interface()); err != nil { // Return error message.返回错误信息
			b.handleErrorString(c, req, err)
			return
		}

		if reqIsValue {
			req = req.Elem()
		}

		bainfo, is := b.preCall(c, obj, req.Interface(), methodName)
		if !is {
			c.JSON(bainfo.RespCode, bainfo.Resp)
			return
		}

		var returnValues []reflect.Value
		returnValues = tvl.Call([]reflect.Value{obj, reflect.ValueOf(apiFun(c)), req})

		if returnValues != nil {
			bainfo.Resp = returnValues[0].Interface()
			rerr := returnValues[1].Interface()
			if rerr != nil {
				bainfo.Error = rerr.(error)
			}

			is = b.afterCall(bainfo, obj)
			if is {
				if bainfo.Error != nil {
					c.JSON(GetResultFunc(RESULT_CODE_ERROR, bainfo.Error.Error(), 0, bainfo.Resp))
				} else {
					c.JSON(GetResultFunc(RESULT_CODE_SUCCESS, "", 0, bainfo.Resp))
				}
			} else {
				if bainfo.Error != nil {
					c.JSON(GetResultFunc(RESULT_CODE_ERROR, bainfo.Error.Error(), 0, bainfo.Resp))
				} else {
					c.JSON(GetResultFunc(RESULT_CODE_ERROR, "", 0, bainfo.Resp))
				}
			}
		}
	}, nil
}

func (b *KApi) handleErrorString(c *gin.Context, req reflect.Value, err error) {
	var fields []string
	if _, ok := err.(validator.ValidationErrors); ok {
		for _, err := range err.(validator.ValidationErrors) {
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

	c.JSON(GetResultFunc(RESULT_CODE_FAIL, fmt.Sprintf("req param : %v", strings.Join(fields, ";")), 0, nil))
	return
}

func (b *KApi) unmarshal(c *gin.Context, v interface{}) error {
	if err := c.ShouldBindHeader(v); err != nil {
		if err != io.EOF {
			if _, ok := err.(validator.ValidationErrors); ok {

			} else {
				fmt.Println("ShouldBindHeader:", err)
				return err
			}

		}

	}
	if err := c.ShouldBindUri(v); err != nil {
		if err != io.EOF {
			if _, ok := err.(validator.ValidationErrors); ok {

			} else {
				fmt.Println("ShouldBindUri:", err)
				return err
			}

		}
	}
	if err := binding.Path.Bind(c, v); err != nil {
		if err != io.EOF {
			if _, ok := err.(validator.ValidationErrors); ok {

			} else {
				fmt.Println("Path.Bind:", err)
				return err
			}

		}
	}

	if err := c.ShouldBindWith(v, binding.Query); err != nil {
		if err != io.EOF {
			if _, ok := err.(validator.ValidationErrors); ok {

			} else {
				fmt.Println("ShouldBindWith.Query:", err)
				return err
			}

		}

	}
	if err := c.ShouldBindJSON(v); err != nil {
		if err != io.EOF {
			fmt.Println("body:", err)
			return err
		}

	}
	return nil
}

func (b *KApi) analysisMethodReqRespByString(Import, objName string, isArray bool, p *goast.Package, modulePkg, moduleFile string) (si *doc.StructInfo) {
	ant := ast.NewStructAnalysis(modulePkg, moduleFile)
	if p == nil {
		bb := ast.EvalSymlinks(modulePkg, moduleFile, Import)
		p, _ = ast.GetAstPackage(Import, bb) // get ast trees.
	}
	si = ant.ParseStruct(p, objName, isArray)

	return
}

// analysisMethodReqResp 解析方法的请求参数或返回参数
func (b *KApi) analysisMethodReqResp(req goast.Expr, imports map[string]string, objPkg string, astPkg *goast.Package, modPkg, modFile string) (si *doc.StructInfo) {
	param := &paramInfo{}
	switch exp := req.(type) {
	case *goast.SelectorExpr: // 非本文件包
		param.Type = exp.Sel.Name
		if x, ok := exp.X.(*goast.Ident); ok {
			param.Import = imports[x.Name]
			param.Pkg = ast.GetImportPkg(param.Import)
		}
	case *goast.StarExpr: // 本文件
		switch expx := exp.X.(type) {
		case *goast.SelectorExpr: // 非本地包
			param.Type = expx.Sel.Name
			if x, ok := expx.X.(*goast.Ident); ok {
				param.Pkg = x.Name
				param.Import = imports[param.Pkg]
			}
		case *goast.Ident: // 本文件
			param.Type = expx.Name
			param.Import = objPkg // 本包
		default:
			//log.ErrorString(fmt.Sprintf("not find any expx.(%v) [%v]", reflect.TypeOf(expx), objPkg))
		}
	case *goast.Ident: // 本文件
		param.Type = exp.Name
		param.Import = objPkg // 本包
	default:
		//log.ErrorString(fmt.Sprintf("not find any exp.(%v) [%v]", reflect.TypeOf(d), objPkg))
	}
	if len(param.Pkg) > 0 {
		var pkg string
		n := strings.LastIndex(param.Import, "/")
		if n > 0 {
			pkg = param.Import[n+1:]
		}
		if len(pkg) > 0 {
			param.Pkg = pkg
		}
	}
	ant := ast.NewStructAnalysis(modPkg, modFile)
	tmp := astPkg
	if len(param.Pkg) > 0 {
		objFile := ast.EvalSymlinks(modPkg, modFile, param.Import)
		tmp, _ = ast.GetAstPackage(param.Pkg, objFile) // get ast trees.
	}
	si = ant.ParseStruct(tmp, param.Type, false)

	return
}

// analysisMethodComment 解析方法注解
func (b *KApi) analysisMethodComment(f *goast.FuncDecl, controllerRoute, objFunc string) (gc *genComment) {
	gc = &genComment{}

	if f.Doc != nil {
		for _, c := range f.Doc.List { // 读取方法的注释
			if prefix, comment, success := internal.GetCommentAfterPrefixRegex(c.Text, objFunc); success {
				switch prefix {
				case "@DEPRECATED":
					gc.IsDeprecated = true
					break
				case "@RESP":
					gc.ResultType = comment
					break
				case "@GET", "@POST", "@PUT", "@DELETE", "@PATCH", "@OPTION", "@HEAD":
					gc.RouterPath = comment
					if controllerRoute != "" {
						gc.RouterPath = controllerRoute + gc.RouterPath
					}

					if len(gc.Methods) > 0 { // 一个方法可以有多个 @HTTPMETHOD 注解
						gc.Methods = append(gc.Methods, strings.ToUpper(strings.TrimPrefix(prefix, "@")))
					} else {
						gc.Methods = []string{strings.ToUpper(strings.TrimPrefix(prefix, "@"))}
					}
					break
				case objFunc:
					gc.Note += comment // 方法注释可以有多个 只要都以方法名开头即可
					break
				}
			}

		}

	}
	return
}

func (b *KApi) analysisController(controller interface{}, model *doc.Model, modPkg string, modFile string) {
	controllerRefVal := reflect.ValueOf(controller)
	internal.Log.Debugf("解析 --> %s", controllerRefVal.Type().String())
	controllerType := reflect.Indirect(controllerRefVal).Type()

	controllerPkgPath := controllerType.PkgPath()

	controllerName := controllerType.Name()

	// find path
	controllerFile := ast.EvalSymlinks(modPkg, modFile, controllerPkgPath)

	controllerAstPkg, _b := ast.GetAstPackage(controllerPkgPath, controllerFile) // get ast trees.
	if _b {
		imports, funMp, cc := ast.AnalysisControllerFile(controllerAstPkg, controllerName)

		refTyp := reflect.TypeOf(controller)
		// 遍历controller方法
		for m := 0; m < refTyp.NumMethod(); m++ {
			method := refTyp.Method(m)
			_, _b := b.checkHandlerFunc(method.Type, true)
			if _b && method.IsExported() {
				if sdl, ok := funMp[method.Name]; ok {
					gc := b.analysisMethodComment(sdl, cc.Route, method.Name)
					if gc != nil {
						checkOnceAdd(controllerName+"/"+method.Name, gc.RouterPath, gc.Methods)
					}
					if b.option.Server.NeedDoc { // output newDoc
						var docReq, docResp *doc.StructInfo
						if sdl.Type.Params.NumFields() > 1 {
							docReq = b.analysisMethodReqResp(sdl.Type.Params.List[1].Type, imports, controllerPkgPath, controllerAstPkg, modPkg, modFile)
						}

						if sdl.Type.Results.NumFields() > 1 {
							docResp = b.analysisMethodReqResp(sdl.Type.Results.List[0].Type, imports, controllerPkgPath, controllerAstPkg, modPkg, modFile)
						} else {
							if gc.ResultType != "" {
								fmt.Println(gc.ResultType)
								aa := strings.Split(gc.ResultType, ".")
								xx := aa[0]
								isArray := strings.HasPrefix(aa[0], "[]")
								if isArray {
									internal.Log.Infof("func:%s result type is array", gc.RouterPath)
									xx = strings.TrimPrefix(aa[0], "[]")
								}

								if len(aa) > 1 {
									if importPath, ok := imports[xx]; ok {
										docResp = b.analysisMethodReqRespByString(importPath, aa[1], isArray, nil, modPkg, modFile)
									}
								} else {
									docResp = b.analysisMethodReqRespByString(controllerPkgPath, gc.ResultType, isArray, controllerAstPkg, modPkg, modFile)
								}

							}
						}
						if gc != nil {
							model.AddOne(cc.TagName, gc.RouterPath, gc.Methods, gc.Note, docReq, docResp, cc.TokenHeader, gc.IsDeprecated)
						}
					}

				}
			}
		}
	}
}

// analysisControllers gen out the Registered config info  by struct object,[prepath + objname.]
func (b *KApi) analysisControllers(router gin.IRoutes, controllers ...interface{}) bool {
	//TODO: 需要解析controller的注释，然后解析controller下方法的注释， 然后解析每个方法的参数, 优化解析性能
	//TODO: groupPath也要加入到文档的路由中
	modPkg, modFile, isFind := ast.GetModuleInfo(2)
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
			model.SetDefinition(b.doc, v1.Req)
			model.SetDefinition(b.doc, v1.Resp)
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
			p.Summary = tagControllerMethod.Note
			p.Description = tagControllerMethod.Note
			url := buildRelativePath(model.Group, tagControllerMethod.RouterPath)

			myreqRef := ""
			p.Parameters = make([]swagger.Element, 0)
			p.Deprecated = tagControllerMethod.IsDeprecated

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
	mp := getInfo()
	for _, c := range cList {
		refTyp := reflect.TypeOf(c)
		refVal := reflect.ValueOf(c)
		t := reflect.Indirect(refVal).Type()
		objName := t.Name()

		// Install the methods
		for m := 0; m < refTyp.NumMethod(); m++ {
			method := refTyp.Method(m)
			_, _b := b.checkHandlerFunc(method.Type, true)
			if _b {
				if v, ok := mp[objName+"/"+method.Name]; ok {
					for _, v1 := range v {
						methods := strings.Join(v1.GenComment.Methods, ",")
						internal.Log.Debugf("%6s  %-20s --> %s", methods, v1.GenComment.RouterPath, t.PkgPath()+".(*"+objName+")."+method.Name)
						_ = b.registerHandlerObj(router, v1.GenComment.Methods, v1.GenComment.RouterPath, method.Name, method.Func, refVal)
					}
				}
			}
		}
	}
	return true
}

// registerHandlerObj 注册路由方法
func (b *KApi) registerHandlerObj(router gin.IRoutes, httpMethod []string, relativePath, methodName string, tvl, obj reflect.Value) error {
	call := b.handlerFuncObj(tvl, obj, methodName)

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
			return fmt.Errorf("请求方式:[%v] 不支持", httpMethod)
		}
	}

	return nil
}
