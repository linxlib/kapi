package kapi

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	binding2 "github.com/gin-gonic/gin/binding"
	"github.com/go-openapi/spec"
	binding3 "github.com/linxlib/binding"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/ast_parser"
	"github.com/linxlib/kapi/internal/comment_parser"
	"io"
	"reflect"
	"strings"
)

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
			rerr := returnValues[1].Interface()

			if rerr != nil {
				c.PureJSON(c.ResultBuilder.OnErrorDetail(rerr.(error).Error(), resp))
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
				internal.Errorf("ShouldBindHeader:%s", err)
				return err
			}
		}
	}

	if err := c.ShouldBindUri(v); err != nil {
		if err != io.EOF {
			if _, ok := binding3.HandleValidationErrors(err); ok {

			} else {
				internal.Errorf("ShouldBindUri: %s", err)
				return err
			}

		}
	}
	if err := binding3.Path.Bind(c.Context, v); err != nil {
		if err != io.EOF {
			if _, ok := binding3.HandleValidationErrors(err); ok {

			} else {
				internal.Errorf("Path.Bind:%s", err)
				return err
			}

		}
	}

	if err := c.ShouldBindWith(v, binding3.Query); err != nil {
		if err != io.EOF {
			if _, ok := binding3.HandleValidationErrors(err); ok {

			} else {
				internal.Errorf("ShouldBindWith.Query:%s", err)
				return err
			}

		}

	}
	if c.ContentType() == "multipart/form-data" {
		if err := c.ShouldBindWith(v, binding2.FormMultipart); err != nil {
			if err != io.EOF {
				if _, ok := binding3.HandleValidationErrors(err); ok {

				} else {
					internal.Errorf("ShouldBindWith.FormMultipart:%s", err)
					return err
				}

			}

		}
	}

	if err := c.ShouldBindJSON(v); err != nil {
		if err != io.EOF {
			internal.Errorf("body EOF:%s", err)
			return err
		}

	}
	return nil
}

func (b *KApi) getStruct(parser *ast_parser.Parser, methodComment *comment_parser.Comment, method *ast_parser.Method, req bool) (s *ast_parser.Struct) {
	if req {
		// has @REQ but request param not defined
		if methodComment.HasReq && len(methodComment.RequestType) > 0 && len(method.Params) <= 1 {
			pkg := ""
			t := ""
			switch len(methodComment.RequestType) {
			case 1:
				pkg = method.PkgPath
				t = methodComment.RequestType[0]
			case 2:
				pkg = methodComment.RequestType[0]
				t = methodComment.RequestType[1]
			}
			f1, _ := parser.Parse(pkg, t)
			s = f1.Structs[0]
		} else if len(method.Params) > 1 {
			s = method.Params[1].Struct
		}
		return
	} else {
		if methodComment.HasResp && len(methodComment.ResultType) > 0 && len(method.Results) <= 0 {
			pkg := ""
			t := ""
			switch len(methodComment.ResultType) {
			case 1:
				pkg = method.PkgPath
				t = methodComment.ResultType[0]
			case 2:
				pkg = methodComment.ResultType[0]
				t = methodComment.ResultType[1]
			}
			f2, _ := parser.Parse(pkg, t)
			s = f2.Structs[0]
		} else if len(method.Results) > 0 {
			s = method.Results[0].Struct
		}
		return
	}

}

func (b *KApi) analysisController(controller interface{}, modPkg string, modFile string) bool {
	controllerRefVal := reflect.ValueOf(controller)
	internal.Debugf("%6s %s", ">", controllerRefVal.Type().String())
	controllerType := reflect.Indirect(controllerRefVal).Type()
	controllerPkgPath := controllerType.PkgPath()
	//parse controller
	parser := ast_parser.NewParser(modPkg, modFile)
	f, err := parser.Parse(controllerPkgPath, controllerType.Name())
	if err != nil {
		internal.Errorf("%+v", err)
		return false
	}
	controllerStruct := f.Structs[0]
	controllerParser := comment_parser.NewParser(controllerStruct.Name, controllerStruct.Docs)
	cp := controllerParser.Parse("")
	//parse methods
	for _, method := range controllerStruct.Methods {
		//only public method will be handled
		if method.Private {
			continue
		}
		//parse method comments
		p := comment_parser.NewParser(method.Name, method.Docs)
		methodComment := p.Parse(cp.Route) //base route

		for m, r := range methodComment.Routes {
			//add routes. which will be registered later
			b.routeInfo.AddFunc(controllerType.Name()+"/"+method.Name, m, r)

			if b.option.Server.NeedDoc {
				if cp.Deprecated {
					methodComment.Deprecated = true //deprecate all method
				}
				var tag = cp.Summary
				if cp.Tag == "" {
					tag = cp.Summary
				}
				//just add tags to swagger
				b.doc.AddTag(spec.NewTag(tag, "", nil))
				sReq := b.getStruct(parser, methodComment, method, true)
				requestParams := b.doc.RequestParams(sReq)
				sResp := b.getStruct(parser, methodComment, method, false)
				responseParams := b.doc.ResponseParams(sResp)

				// 方法可能注册为多条路由
				for r, m := range methodComment.Routes {
					if strings.Contains(r, "{") || strings.Contains(r, "}") {
						internal.Errorf("[%s]path route {path} not supported. use :path instead", r)
						return false
					}

					b.doc.AddRoute(m, r,
						methodComment.Deprecated,
						methodComment.GetDescription(","),
						tag,
						requestParams,
						responseParams)
				}
			}
		}

	}
	return true
}

// analysisControllers
func (b *KApi) analysisControllers(controllers ...interface{}) bool {
	defer internal.Spend("analysisControllers")()
	internal.Debugf("analysis controllers...")
	modPkg, modFile, isFind := internal.GetModuleInfo(2)
	if !isFind {
		return false
	}
	for _, c := range controllers {
		if !b.analysisController(c, modPkg, modFile) {
			return false
		}
	}
	return true
}

// register 注册路由到gin
func (b *KApi) register(cList ...interface{}) bool {
	defer internal.Spend("register routes")()
	internal.Debugf("register controllers..")
	mp := b.routeInfo.GetGenInfo().Routes
	for _, c := range cList {
		refTyp := reflect.TypeOf(c)
		refVal := reflect.ValueOf(c)
		t := reflect.Indirect(refVal).Type()
		objName := t.Name()
		err := b.Apply(c)
		if err != nil {
			internal.Errorf("%+v", err)
			return false
		}
		// Install the Method
		for m := 0; m < refTyp.NumMethod(); m++ {
			method := refTyp.Method(m)
			if !method.IsExported() {
				continue
			}
			k := objName + "/" + method.Name
			for _, item := range mp {
				if item.Key == k {
					internal.Debugf("%6s  %-30s --> %s", item.Method, item.RouterPath, t.PkgPath()+".(*"+objName+")."+method.Name)
					err := b.registerMethodToRouter(item.Method,
						item.RouterPath,
						refVal.Interface(),
						refVal.Method(m).Interface())
					if err != nil {
						internal.Errorf("%s", err)
						return false
					}
				}
			}

		}
	}
	return true
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
func (b *KApi) registerMethodToRouter(httpMethod string, relativePath string, controller, method interface{}) error {
	call := b.handle(controller, method)
	switch strings.ToUpper(httpMethod) {
	case "POST":
		b.engine.POST(relativePath, call)
	case "GET":
		b.engine.GET(relativePath, call)
	case "DELETE":
		b.engine.DELETE(relativePath, call)
	case "PATCH":
		b.engine.PATCH(relativePath, call)
	case "PUT":
		b.engine.PUT(relativePath, call)
	case "OPTIONS":
		b.engine.OPTIONS(relativePath, call)
	case "HEAD":
		b.engine.HEAD(relativePath, call)
	case "ANY":
		b.engine.Any(relativePath, call)
	default:
		return fmt.Errorf("http method:[%v --> %s] not supported", httpMethod, relativePath)
	}

	return nil
}

// genRouterCode 生成gen.gob
func (b *KApi) genRouterCode() {
	defer internal.Spend("generate router code")()
	if b.doc == nil {
		return
	}
	b.routeInfo.SetApiBody(b.doc)
	go b.routeInfo.WriteOut()
}
