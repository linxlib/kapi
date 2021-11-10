package kapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gitee.com/kirile/kapi/ast"
	"gitee.com/kirile/kapi/binding"
	"gitee.com/kirile/kapi/doc"
	"gitee.com/kirile/kapi/doc/swagger"
	"gitee.com/kirile/kapi/internal"
	goast "go/ast"
	"io"
	"net/http"
	"reflect"
	"regexp"
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
	num := typ.NumIn() - offset
	if num == 1 || num == 2 { // 参数检查
		ctxType := typ.In(0 + offset)

		// go-gin 框架的默认方法
		if ctxType == reflect.TypeOf(&gin.Context{}) {
			return num, true
		}

		// Customized context . 自定义的context
		if ctxType == b.customContextType {
			return num, true
		}

		// maybe interface
		if b.customContextType.ConvertibleTo(ctxType) {
			return num, true
		}

	}
	return num, false
}

// HandlerFunc Get and filter the parameters to be bound (object call type)
func (b *KApi) handlerFuncObj(tvl, obj reflect.Value, methodName string) gin.HandlerFunc { // 获取并过滤要绑定的参数(obj 对象类型)
	typ := tvl.Type()
	if typ.NumIn() == 2 { //1个参数的方法
		ctxType := typ.In(1)

		apiFun := func(c *gin.Context) interface{} { return c }
		if ctxType == b.customContextType { // Customized context . 自定义的context
			apiFun = b.apiFun
		} else if !(ctxType == reflect.TypeOf(&gin.Context{})) {
			_log.Errorln("方法不受支持 " + runtime.FuncForPC(tvl.Pointer()).Name() + " !")
		}

		return func(c *gin.Context) {
			//加入recover处理
			defer func() {
				if err := recover(); err != nil {
					b.option.recoverErrorFunc(err)
				}
			}()

			bainfo, is := b.beforeCall(c, obj, nil, methodName)
			if !is {
				c.JSON(http.StatusBadRequest, bainfo.Resp)
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
						c.JSON(DefaultGetResult(RESULT_CODE_ERROR, bainfo.Error.Error(), 0, bainfo.Resp))
					} else {
						c.JSON(DefaultGetResult(RESULT_CODE_SUCCESS, "", 0, bainfo.Resp))
					}
				} else {
					if bainfo.Error != nil {
						c.JSON(DefaultGetResult(RESULT_CODE_ERROR, bainfo.Error.Error(), 0, bainfo.Resp))
					} else {
						c.JSON(DefaultGetResult(RESULT_CODE_ERROR, "", 0, bainfo.Resp))
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

//beforeCall 调用前处理
func (b *KApi) beforeCall(c *gin.Context, obj reflect.Value, req interface{}, methodName string) (*InterceptorContext, bool) {
	info := &InterceptorContext{
		C:        &Context{c},
		FuncName: fmt.Sprintf("%v.%v", reflect.Indirect(obj).Type().Name(), methodName), // 函数名
		Req:      req,                                                                   // 调用前的请求参数
		Context:  context.Background(),                                                  // 占位参数，可用于存储其他参数，前后连接可用
	}
	//处理控制器的拦截器
	is := true
	if bfobj, ok := obj.Interface().(Interceptor); ok { // 本类型
		is = bfobj.GinBefore(info)
	}
	//处理全局的拦截器
	if is && b.beforeAfter != nil {
		is = b.beforeAfter.GinBefore(info)
	}
	return info, is
}

//afterCall 调用后处理
func (b *KApi) afterCall(info *InterceptorContext, obj reflect.Value) bool {
	is := true
	if bfobj, ok := obj.Interface().(Interceptor); ok { // 本类型
		is = bfobj.GinAfter(info)
	}
	if is && b.beforeAfter != nil {
		is = b.beforeAfter.GinAfter(info)
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

		bainfo, is := b.beforeCall(c, obj, req.Interface(), methodName)
		if !is {
			c.JSON(http.StatusBadRequest, bainfo.Resp)
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
					c.JSON(DefaultGetResult(RESULT_CODE_ERROR, bainfo.Error.Error(), 0, bainfo.Resp))
				} else {
					c.JSON(DefaultGetResult(RESULT_CODE_SUCCESS, "", 0, bainfo.Resp))
				}
			} else {
				if bainfo.Error != nil {
					c.JSON(DefaultGetResult(RESULT_CODE_ERROR, bainfo.Error.Error(), 0, bainfo.Resp))
				} else {
					c.JSON(DefaultGetResult(RESULT_CODE_ERROR, "", 0, bainfo.Resp))
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

	c.JSON(DefaultGetResult(RESULT_CODE_FAIL, fmt.Sprintf("req param : %v", strings.Join(fields, ";")), 0, nil))
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

func (b *KApi) parseStruct(req, resp *paramInfo, astPkg *goast.Package, modPkg, modFile string) (r, p *doc.StructInfo) {
	ant := ast.NewStructAnalysis(modPkg, modFile)
	if req != nil {
		tmp := astPkg
		if len(req.Pkg) > 0 {
			objFile := ast.EvalSymlinks(modPkg, modFile, req.Import)
			tmp, _ = ast.GetAstPkgs(req.Pkg, objFile) // get ast trees.
		}
		r = ant.ParseStruct(tmp, req.Type)
	}

	if resp != nil {
		tmp := astPkg
		if len(resp.Pkg) > 0 {
			objFile := ast.EvalSymlinks(modPkg, modFile, resp.Import)
			tmp, _ = ast.GetAstPkgs(resp.Pkg, objFile) // get ast trees.
		}
		p = ant.ParseStruct(tmp, resp.Type)
	}

	return
}

func analysisParam(f *goast.FieldList, imports map[string]string, objPkg string, n int) (param *paramInfo) {
	if f != nil {
		if f.NumFields() > 1 {
			param = &paramInfo{}
			d := f.List[n].Type
			switch exp := d.(type) {
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
		}
	}

	if param != nil {
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
	}
	return
}

var routeRegex = regexp.MustCompile(`(@\w+)\s+(\S+)`)
var deprecatedRegex = regexp.MustCompile(`(@\w+)`)

func (b *KApi) parseComments(f *goast.FuncDecl, controllerRoute, objFunc string, imports map[string]string, objPkg string, num int) (bool, []genComment, *paramInfo, *paramInfo) {
	var note string
	var isDeprecated = false
	var gcs []genComment
	req := analysisParam(f.Type.Params, imports, objPkg, 1)   // 第二个参数作为请求参数
	resp := analysisParam(f.Type.Results, imports, objPkg, 0) // 第一个返回值作为resp

	if f.Doc != nil {
		for _, c := range f.Doc.List { // 读取方法的注释
			gc := genComment{}
			t := strings.TrimSpace(strings.TrimPrefix(c.Text, "//")) //去掉注释前缀并移除首尾的空格
			if strings.HasPrefix(t, "@") {                           //以
				matches := routeRegex.FindStringSubmatch(t)
				if len(matches) == 3 { // 第一个是自身全部
					gc.RouterPath = matches[2]
					if controllerRoute != "" {
						gc.RouterPath = controllerRoute + gc.RouterPath
					}
					if strings.Contains(gc.RouterPath, "/") {
						gc.RouterPath = strings.TrimSuffix(gc.RouterPath, "/")
					}
					methods := matches[1]
					if methods == "" {
						gc.Methods = []string{"get"}
					} else {
						gc.Methods = []string{strings.ToLower(strings.TrimPrefix(methods, "@"))}
					}
					gcs = append(gcs, gc)
				} else {
					//处理其他注释
					matches1 := deprecatedRegex.FindStringSubmatch(t)

					if len(matches1) == 2 && matches1[1] == "@DEPRECATED" {
						isDeprecated = true
					}
				}

			} else if strings.HasPrefix(t, objFunc) { // 以方法名开头的注释 设置为api的描述
				t = strings.TrimSpace(strings.TrimPrefix(t, objFunc))
				note += t
			}
		}

	}

	//default
	if len(gcs) == 0 {
		return isDeprecated, make([]genComment, 0), nil, nil
	}

	// add note 添加注释
	for i := 0; i < len(gcs); i++ {
		gcs[i].Note = note
	}

	return isDeprecated, gcs, req, resp
}

// tryGenRegister gen out the Registered config info  by struct object,[prepath + objname.]
func (b *KApi) tryGenRegister(router gin.IRoutes, controllers ...interface{}) bool {
	//TODO: 需要解析controller的注释，然后解析controller下方法的注释， 然后解析每个方法的参数

	modPkg, modFile, isFind := ast.GetModuleInfo(2)
	if !isFind {
		return false
	}

	groupPath := b.BasePath(router)
	doc := doc.NewDoc(groupPath)
	for _, c := range controllers {
		refVal := reflect.ValueOf(c)
		_log.Debugf("解析 --> %s", refVal.Type().String())
		t := reflect.Indirect(refVal).Type()

		objPkg := t.PkgPath()

		objName := t.Name()
		tagName := objName
		route := ""
		tokenHeader := ""

		// find path
		objFile := ast.EvalSymlinks(modPkg, modFile, objPkg)

		astPkgs, _b := ast.GetAstPkgs(objPkg, objFile) // get ast trees.
		if _b {
			imports := ast.AnalysisImport(astPkgs)
			funMp := ast.GetObjFunMp(astPkgs, objName)
			t, r, th := ast.AnalysisControllerComments(astPkgs, objName)
			if t != "" {
				tagName = t
			}
			if r != "" {
				route = r
			}
			if th != "" {
				tokenHeader = th
			}

			refTyp := reflect.TypeOf(c)
			// Install the methods
			for m := 0; m < refTyp.NumMethod(); m++ {
				method := refTyp.Method(m)
				num, _b := b.checkHandlerFunc(method.Type, true)
				if _b {
					if sdl, ok := funMp[method.Name]; ok {
						isDeprecated, gcs, req, resp := b.parseComments(sdl, route, method.Name, imports, objPkg, num)
						if b.option.needDoc { // output doc
							docReq, docResp := b.parseStruct(req, resp, astPkgs, modPkg, modFile)
							for _, gc := range gcs {
								doc.AddOne(tagName, gc.RouterPath, gc.Methods, gc.Note, docReq, docResp, tokenHeader, isDeprecated)
								checkOnceAdd(objName+"/"+method.Name, gc.RouterPath, gc.Methods)
							}
						} else {
							for _, gc := range gcs {
								checkOnceAdd(objName+"/"+method.Name, gc.RouterPath, gc.Methods)
							}
						}

					}
				}
			}
		}
	}

	if b.option.needDoc {
		b.addDocModel(doc)
	}
	return true
}

func (b *KApi) addDocModel(model *doc.Model) {
	var sortStr []string
	for k, v := range model.MP {
		for _, v1 := range v {
			model.SetDefinition(b.doc, v1.Req)
			model.SetDefinition(b.doc, v1.Resp)
		}
		sortStr = append(sortStr, k)
	}
	sort.Strings(sortStr)

	for _, k := range sortStr {
		v := model.MP[k]
		tag := swagger.Tag{Name: k}
		b.doc.AddTag(tag)
		for _, v1 := range v {
			var p swagger.Param
			p.Tags = []string{k}
			p.Summary = v1.Note
			p.Description = v1.Note
			p.OperationID = v1.Methods[0] + "_" + strings.ReplaceAll(v1.RouterPath, "/", "_")
			myreqRef := ""
			p.Parameters = make([]swagger.Element, 0)
			p.Deprecated = v1.IsDeprecated

			if v1.Req != nil {
				for _, item := range v1.Req.Items {
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
						p.Parameters = append(p.Parameters, swagger.Element{
							In:          "formData",
							Name:        item.Name,
							Description: item.Note,
							Required:    item.Required,
							Type:        swagger.GetKvType(item.Type, false, true),
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
						myreqRef = "#/definitions/" + v1.Req.Name
						exist := false
						for _, parameter := range p.Parameters {
							if parameter.Name == v1.Req.Name {
								exist = true
							}
						}
						if !exist {
							p.Parameters = append(p.Parameters, swagger.Element{
								In:          "body",
								Name:        v1.Req.Name,
								Description: item.Note,
								Required:    true,
								Schema: &swagger.Schema{
									Ref: myreqRef,
								},
							})
						}

					}

				}

			} else {

			}
			if v1.TokenHeader != "" {
				p.Parameters = append(p.Parameters, swagger.Element{
					In:          "header",
					Name:        v1.TokenHeader,
					Description: v1.TokenHeader,
					Required:    true,
					Type:        "string",
					Schema:      nil,
					Default:     "",
				})
			}
			if v1.Resp != nil {
				p.Responses = make(map[string]swagger.Resp)
				for _, item := range v1.Resp.Items {
					if item.IsArray {
						p.Responses["200"] = swagger.Resp{
							Description: "successful result",
							Schema: map[string]interface{}{
								"type": "array",
								"items": map[string]string{
									"$ref": "#/definitions/" + v1.Resp.Name,
								},
							},
						}
					} else {
						p.Responses["200"] = swagger.Resp{
							Description: "成功返回",
							Schema: map[string]interface{}{
								"$ref": "#/definitions/" + v1.Resp.Name,
							},
						}
					}
				}
			}

			b.doc.AddPatch(buildRelativePath(model.Group, v1.RouterPath), p, v1.Methods...)
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
						_log.Debugf("%6s  %-20s --> %s", methods, v1.GenComment.RouterPath, t.PkgPath()+".(*"+objName+")."+method.Name)
						b.registerHandlerObj(router, v1.GenComment.Methods, v1.GenComment.RouterPath, method.Name, method.Func, refVal)
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
