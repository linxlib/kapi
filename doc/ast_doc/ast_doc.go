package ast_doc

import (
	"fmt"
	"github.com/linxlib/kapi/doc"
	"github.com/linxlib/kapi/internal"
	"go/ast"
	"go/token"
	"strings"
)

var importFile = make(map[string]string) // 自定义包文件

type ctlScheme struct {
	Imports     map[string]string
	FuncMap     map[string]*ast.FuncDecl
	TagName     string //@TAG
	Route       string //@ROUTE
	TokenHeader string //@AUTH
}

var (
	getAstPackagesCache = make(map[string]*ast.Package)
	parseStructCache    = make(map[string]*doc.StructInfo)
)

// AddImportFile 添加自定义import文件列表
func AddImportFile(k, v string) {
	importFile[k] = v
}

type methodComment struct {
	RouterPath   string
	IsDeprecated bool
	ResultType   string
	Summary      string //方法说明
	Description  string // 方法注释
	Methods      []string
	TokenHeader  string
}

type AstDoc struct {
	astPkg            *ast.Package
	modPkg            string
	modFile           string
	controllerScheme  *ctlScheme
	controllerPkgPath string
}

func NewAstDoc(modPkg string, modFile string) *AstDoc {
	return &AstDoc{
		astPkg:  nil,
		modPkg:  modPkg,
		modFile: modFile,
	}
}
func (a *AstDoc) FillPackage(objPkg string) error {
	a.controllerPkgPath = objPkg
	f := a.getFileByPkgPath(objPkg)
	var b bool
	a.astPkg, b = GetAstPackage(objPkg, f)
	if !b {
		return fmt.Errorf("cannot get ast package of %s", objPkg)
	}
	return nil
}

func (a *AstDoc) getFileByPkgPath(objPkg string) string {
	if strings.EqualFold(objPkg, "main") { // if main return default path
		return a.modFile
	}

	if strings.HasPrefix(objPkg, a.modPkg) {
		return a.modFile + strings.Replace(objPkg[len(a.modPkg):], ".", "/", -1)
	}

	// 自定义文件中查找
	tmp := importFile[objPkg]
	if len(tmp) > 0 {
		return tmp
	}

	// get the error space
	panic(fmt.Errorf("can not eval pkg:[%v] must include [%v]", objPkg, a.modPkg))
}

func (a *AstDoc) ResolveController(controllerName string) *ctlScheme {
	comment := &ctlScheme{
		Imports: make(map[string]string),
		FuncMap: make(map[string]*ast.FuncDecl),
		TagName: controllerName,
	}

	for _, f := range a.astPkg.Files {
		for _, p := range f.Imports {
			handleImportSpec(p, comment.Imports)
		}
		for _, d := range f.Decls {
			switch specDecl := d.(type) {
			case *ast.FuncDecl:
				if specDecl.Recv != nil {
					if exp, ok := specDecl.Recv.List[0].Type.(*ast.StarExpr); ok { // Check that the type is correct first before throwing to parser
						if strings.Compare(fmt.Sprint(exp.X), controllerName) == 0 { // is the same struct
							comment.FuncMap[specDecl.Name.String()] = specDecl // catch
						}
					}
				}
			case *ast.GenDecl:
				for _, spec := range specDecl.Specs {
					switch specDecl.Tok {
					case token.TYPE:
						spec := spec.(*ast.TypeSpec)
						switch spec.Type.(type) {
						case *ast.StructType:
							if spec.Name.Name == controllerName { // find it
								if specDecl.Doc != nil { // 如果有注释
									for _, v := range specDecl.Doc.List { // 结构体注释
										if prefix, commentContent, b := internal.GetCommentAfterPrefixRegex(v.Text, controllerName); b {
											switch prefix {
											case "@TAG":
												comment.TagName = commentContent
											case "@ROUTE":
												comment.Route = commentContent
											case "@AUTH":
												if commentContent == "" {
													commentContent = "Authorization"
												}
												comment.TokenHeader = commentContent
											case controllerName:
												comment.TagName = commentContent
											default:
											}
										}
									}

								}

							}
						}
					}
				}
			}
		}
	}
	a.controllerScheme = comment
	return comment
}

func (a *AstDoc) ResolveMethod(methodName string) (*methodComment, *doc.StructInfo, *doc.StructInfo) {
	sdl, gc := a.resolveMethodComment(methodName)
	var docReq, docResp *doc.StructInfo
	if sdl != nil && sdl.Type != nil && sdl.Type.Params != nil && sdl.Type.Params.NumFields() > 1 {
		docReq = a.resolveMethodReqResp(sdl.Type.Params.List[1].Type)
	}
	if sdl != nil && sdl.Type != nil && sdl.Type.Results != nil && sdl.Type.Results.NumFields() > 1 {
		docResp = a.resolveMethodReqResp(sdl.Type.Results.List[0].Type)
	} else {
		if gc != nil && gc.ResultType != "" {
			docResp = a.resolveMethodRespByString(gc.ResultType)
		}
	}
	return gc, docReq, docResp
}

func (a *AstDoc) resolveMethodComment(methodName string) (*ast.FuncDecl, *methodComment) {
	if f, ok := a.controllerScheme.FuncMap[methodName]; ok {
		gc := &methodComment{}

		if f.Doc != nil {
			for _, c := range f.Doc.List { // comment list
				if prefix, comment, success := internal.GetCommentAfterPrefixRegex(c.Text, methodName); success {
					switch prefix {
					case "@DEPRECATED":
						gc.IsDeprecated = true
						break
					case "@RESP":
						gc.ResultType = comment
						break
					case "@DESC":
						gc.Description += comment + "\n" //we can have multiple @DESC to multiline description
					case "@GET", "@POST", "@PUT", "@DELETE", "@PATCH", "@OPTION", "@HEAD":
						gc.RouterPath = comment
						if a.controllerScheme.Route != "" {
							gc.RouterPath = a.controllerScheme.Route + gc.RouterPath
							gc.RouterPath = strings.TrimSuffix(gc.RouterPath, "/")
						}

						if len(gc.Methods) > 0 { //we can also have multiple @HTTPMETHOD
							gc.Methods = append(gc.Methods, strings.ToUpper(strings.TrimPrefix(prefix, "@")))
						} else {
							gc.Methods = []string{strings.ToUpper(strings.TrimPrefix(prefix, "@"))}
						}
						break
					case methodName: //if prefix is equal to method name
						gc.Summary = comment // summary can have only one
						break
					}
				}

			}

		}
		return f, gc
	}
	return nil, nil

}

func (a *AstDoc) resolveMethodReqResp(req ast.Expr) *doc.StructInfo {
	// paramInfo 参数类型描述
	type paramInfo struct {
		Pkg    string // 包名
		Type   string // 类型
		Import string // import 包
	}
	param := &paramInfo{}
	switch exp := req.(type) {
	case *ast.SelectorExpr: // struct not in current package
		param.Type = exp.Sel.Name
		if x, ok := exp.X.(*ast.Ident); ok {
			param.Import = a.controllerScheme.Imports[x.Name] // get import by package name
			param.Pkg = GetImportPkg(param.Import)
		}
	case *ast.StarExpr: // current package
		switch expx := exp.X.(type) {
		case *ast.SelectorExpr: // 非本地包
			param.Type = expx.Sel.Name
			if x, ok := expx.X.(*ast.Ident); ok {
				param.Pkg = x.Name
				param.Import = a.controllerScheme.Imports[param.Pkg]
			}
		case *ast.Ident: // 本文件
			param.Type = expx.Name
			param.Import = a.controllerPkgPath // 本包
		default:
			//log.ErrorString(fmt.Sprintf("not find any expx.(%v) [%v]", reflect.TypeOf(expx), objPkg))
		}
	case *ast.Ident: // current file
		param.Type = exp.Name
		param.Import = a.controllerPkgPath // current package
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
	ant := NewStructAnalysis(a.modPkg, a.modFile)
	tmp := a.astPkg
	if len(param.Pkg) > 0 {
		objFile := a.getFileByPkgPath(param.Import)
		tmp, _ = GetAstPackage(param.Pkg, objFile) // get ast trees.
	}
	return ant.ParseStruct(tmp, param.Type, false)
}

func (a *AstDoc) resolveMethodRespByString(resultType string) *doc.StructInfo {
	aa := strings.Split(resultType, ".")
	xx := aa[0]
	isArray := strings.HasPrefix(aa[0], "[]")
	if isArray {
		xx = strings.TrimPrefix(aa[0], "[]")
	}
	if len(aa) > 1 {
		if importPath, ok := a.controllerScheme.Imports[xx]; ok {
			ant := NewStructAnalysis(a.modPkg, a.modFile)
			bb := a.getFileByPkgPath(importPath)
			p, _ := GetAstPackage(importPath, bb) // get ast trees.
			return ant.ParseStruct(p, aa[1], isArray)
		} else {
			return nil
		}
	} else {
		ant := NewStructAnalysis(a.modPkg, a.modFile)
		return ant.ParseStruct(a.astPkg, xx, isArray)
	}

}
func init() {
	AddImportFile("mime/multipart", "mime/multipart")
}
