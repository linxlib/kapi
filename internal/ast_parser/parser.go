// Package ast_parser get type information from a package using go/ast
package ast_parser

import (
	"errors"
	"fmt"
	"github.com/linxlib/kapi/internal"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"sync"
)

type ParserLogger interface {
	Error(...any)
	Info(...any)
	Infof(string, ...any)
}

type _logger struct {
}

func (_ _logger) Error(a ...any) {
	//log.Println(a...)
}

func (_ _logger) Info(a ...any) {
	//log.Println(a...)
}

func (_ _logger) Infof(s string, a ...any) {
	//log.Printf(s+"\n", a...)
}

type Parser struct {
	dir         string                   //当前结构所在目录
	pkg         string                   //当前结构所在包
	modFile     string                   //项目的模块文件
	modPkg      string                   //项目的模块包
	astPackages map[string]*ast.Package  //所有解析过的ast package
	imports     map[string][]*importItem //包对应的所有导入信息，每个包一组，即包下所有文件的导入会在一组中
	lock        sync.RWMutex
	ignoreList  map[string]string
	logger      ParserLogger
}

// getPkgDir returns the absolute directory path of a package
//
//	@param objPkg
//
//	@return string
func (p *Parser) getPkgDir(objPkg string) string {
	if strings.EqualFold(objPkg, "main") { // if main return default path
		return p.modFile
	}
	if strings.HasPrefix(objPkg, p.modPkg) {
		return p.modFile + strings.Replace(objPkg[len(p.modPkg):], ".", "/", -1)
	}

	//// 自定义文件中查找
	//tmp := importFile[objPkg]
	//if len(tmp) > 0 {
	//	return tmp
	//}
	return ""
	// get the error space
	//panic(fmt.Errorf("can not eval pkg:[%v] must include [%v]", objPkg, a.modPkg))
}

func NewParser(modPkg, modFile string) (p *Parser) {
	p = &Parser{
		modPkg:      modPkg,
		modFile:     modFile,
		pkg:         "",
		dir:         "",
		astPackages: make(map[string]*ast.Package),
		ignoreList: map[string]string{
			"github.com/linxlib/kapi": "Context",
		},
		logger: &_logger{},
	}
	return p
}

// Ignore a type by specified import path and type name
//
//	@param importPath
//	@param t type name
func (p *Parser) Ignore(importPath, t string) {
	if _, ok := p.ignoreList[importPath]; ok {
		return
	}
	p.ignoreList[importPath] = t
}

// parseField 递归解析结构体字段
//
//	@param importPath 导入路径
//	@param importName 导入名
//	@param fieldType 字段类型
//func (p *Parser) parseField(importPath, importName, fieldType string) []Field {
//
//}

// handleImportSpec get all imports map from the package directory
//
//	@param is
//	@param imports
//
//	@return []*importItem
func (p *Parser) handleImportSpec(is *ast.ImportSpec, imports []*importItem) []*importItem {
	k := ""
	if is.Name != nil {
		k = is.Name.Name
	}
	v := strings.Trim(is.Path.Value, `"`)
	if len(k) == 0 {
		n := strings.LastIndex(v, "/")
		if n > 0 {
			k = v[n+1:]
		} else {
			k = v
		}
	}
	imports = append(imports, &importItem{
		name:       k,
		importPath: v,
	})
	return imports
}

var errIgnored = errors.New("ignored")

// handleParam parse method's param
//
//	@param v
//	@param importMP
//
//	@return *Param
//	@return error
func (p *Parser) handleParam(v *ast.Field, importMP []*importItem) (*Param, error) {
	a, isSlice, isPointer, err := getType(v.Type)
	if err != nil {
		return nil, err
	}
	n := ""
	if v.Names != nil {
		n = v.Names[0].Name
	}
	param := &Param{
		Name:       n,
		typeString: a,
		PkgPath:    p.pkg,
		Pointer:    isPointer,
		Slice:      isSlice,
		Type:       strings.Trim(a, "*"),
	}
	i1, isbuiltin, t1 := getImportAndType(a, param.PkgPath, importMP)
	if !isbuiltin {
		f1, err := p.Parse(i1, t1)
		if err == nil {
			param.Struct = f1.Structs[0]
		} else {
			if errors.Is(err, errIgnored) {
				param.ignoreParse = true
			} else {
				p.logger.Error(err)
			}
		}
	} else {
		param.ignoreParse = true
	}
	return param, nil
}

// Parse a type in a package
//
//	@param pkg package
//	@param s type name
//
//	@return error
func (p *Parser) Parse(pkg, s string) (*File, error) {
	//TODO: 内置类型的解析
	if internal.IsInternalType(pkg) {
		f := &File{
			Name:    "",
			Imports: nil,
			PkgPath: "",
			Structs: []*Struct{
				&Struct{
					Name:    pkg,
					PkgPath: "",
					Fields:  nil,
					Methods: nil,
					Docs:    nil,
				},
			},
			Docs: nil,
		}
		return f, nil
	}
	//一些类型不进行解析
	if a, ok := p.ignoreList[pkg]; ok && a == s {
		return nil, errIgnored
	}
	p.logger.Infof("Parse %s %s", pkg, s)
	//获取包和目录
	p.pkg = pkg
	p.dir = p.getPkgDir(pkg)
	k := p.pkg + "_" + p.dir
	//这里作为缓存，相同包下不会重复解析
	var apkg *ast.Package
	if tmpAstPkg, ok := p.astPackages[k]; !ok {
		p.logger.Info(s)
		//获取包下所有的ast package
		dir, err := parser.ParseDir(token.NewFileSet(), p.dir, nil, parser.ParseComments|parser.AllErrors|parser.DeclarationErrors)
		if err != nil {
			return nil, nil
		}
		i := strings.LastIndex(p.pkg, "/")
		p.astPackages[k] = dir[p.pkg[i+1:]]
		apkg = p.astPackages[k]
	} else {
		p.logger.Info(s, "[cache]")
		apkg = tmpAstPkg
	}
	//相同包不会多次处理
	var importMP []*importItem
	if tmpImportMP, ok := p.imports[p.pkg]; !ok {
		for _, f := range apkg.Files {
			for _, p1 := range f.Imports {
				importMP = p.handleImportSpec(p1, importMP)
			}
		}
	} else {
		importMP = tmpImportMP
	}
	f := new(File)
	tmp := doc.New(apkg, "", doc.AllDecls|doc.AllMethods)
	for _, t := range tmp.Types {
		if t == nil || t.Decl == nil {
			return nil, errors.New("t or t.Decl is nil")
		}
		if t.Name != s {
			continue
		}

		for _, spec := range t.Decl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				return nil, errors.New("not a *ast.TypeSpec")
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return nil, errors.New("not a *ast.StructType")
			}

			f.Name = tmp.Name
			f.Imports = importMP
			f.PkgPath = p.pkg

			parsedStruct := &Struct{
				Name:    t.Name,
				PkgPath: p.pkg,
				Fields:  make([]*Field, 0, len(structType.Fields.List)),
				Docs:    getDocsForStruct(t.Doc), //结构体注释
				Methods: make([]*Method, 0),
			}
			for _, fvalue := range structType.Fields.List {
				name := ""
				if len(fvalue.Names) > 0 {
					name = fvalue.Names[0].Obj.Name
				}
				field := &Field{
					Name:    name,
					PkgPath: p.pkg,
					Type:    "",
					Tag:     "",
					Pointer: false,
					Slice:   false,
				}
				if len(field.Name) > 0 {
					field.Private = strings.ToLower(string(field.Name[0])) == string(field.Name[0])
				}

				if fvalue.Doc != nil {
					field.Docs = getDocsForField(fvalue.Doc)
				}
				if fvalue.Comment != nil {
					field.Comment = cleanDocText(fvalue.Comment.Text())
				}
				if fvalue.Tag != nil {
					field.Tag = reflect.StructTag(strings.Trim(fvalue.Tag.Value, "`"))
					field.hasTag = true
				}
				var err error
				field.Type, field.Slice, field.Pointer, err = getType(fvalue.Type)
				if err != nil {
					return nil, err
				}

				parsedStruct.Fields = append(parsedStruct.Fields, field)
			}
			f.Structs = append(f.Structs, parsedStruct)
		}
		//结构体方法
		for _, spec := range t.Methods {
			funcDecl := spec.Decl

			receiver, _, isPointer, _ := getType(funcDecl.Recv.List[0].Type)
			method := &Method{
				Name:    funcDecl.Name.Name,
				PkgPath: p.pkg,
				Receiver: &Receiver{
					Name:    funcDecl.Recv.List[0].Names[0].Name,
					Pointer: isPointer,
					Type:    receiver,
				},
				Private: strings.ToLower(string(funcDecl.Name.Name[0])) == string(funcDecl.Name.Name[0]),
				Params:  []*Param{},
				Results: []*Param{},
				Docs:    getDocsForStruct(spec.Doc),
			}

			//参数
			var tmpArgs []string
			for _, v := range funcDecl.Type.Params.List {
				param, err := p.handleParam(v, importMP)
				if err != nil {
					return nil, err
				}

				method.Params = append(method.Params, param)

				var tmpNames []string
				for _, n := range v.Names {
					tmpNames = append(tmpNames, n.Name)
				}
				tmpArgs = append(tmpArgs, strings.Join(tmpNames, ", ")+" "+param.typeString)
			}
			//返回值
			var tmpReturns []string
			if funcDecl != nil && funcDecl.Type != nil && funcDecl.Type.Results != nil && funcDecl.Type.Results.List != nil {
				for _, v := range funcDecl.Type.Results.List {
					param, err := p.handleParam(v, importMP)
					if err != nil {
						return nil, err
					}
					method.Results = append(method.Results, param)
					var tmpNames []string
					for _, n := range v.Names {
						tmpNames = append(tmpNames, n.Name)
					}
					tmpReturns = append(tmpReturns, strings.Join(tmpNames, ", ")+" "+param.typeString)
				}
			}
			method.Signature = method.Name + "(" + strings.Join(tmpArgs, ", ") + ") (" + strings.Join(tmpReturns, ", ") + ")"

			// find struct and add method
			for k, v := range f.Structs {
				tmp := strings.Trim(method.Receiver.Type, "*")
				if v.Name == tmp {
					f.Structs[k].Methods = append(f.Structs[k].Methods, method)
				}
			}
		}
	}
	return f, nil
}

func getDocsForStruct(doc string) []string {
	if doc == "" {
		return []string{}
	}
	trimmed := strings.Trim(doc, "\n")
	if trimmed == "" {
		return []string{}
	}
	tmp := strings.Split(trimmed, "\n")

	docs := make([]string, 0, len(tmp))
	for _, v := range tmp {
		docs = append(docs, cleanDocText(v))
	}
	return docs
}

func getDocsForField(cg *ast.CommentGroup) []string {
	if cg == nil {
		return []string{}
	}
	docs := make([]string, 0, len(cg.List))
	for _, v := range cg.List {
		docs = append(docs, cleanDocText(v.Text))
	}
	return docs
}
func cleanDocText(doc string) string {
	reverseString := func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	}

	if strings.HasPrefix(doc, "// ") {
		doc = strings.Replace(doc, "// ", "", 1)
	} else if strings.HasPrefix(doc, "//") {
		doc = strings.Replace(doc, "//", "", 1)
	} else if strings.HasPrefix(doc, "/*") {
		doc = strings.Replace(doc, "/*", "", 1)
	}
	if strings.HasSuffix(doc, "*/") {
		doc = reverseString(strings.Replace(reverseString(doc), "/*", "", 1))
	}
	return strings.Trim(strings.Trim(doc, " "), "\n")
}
func justTypeString(a string, b, c bool, err error) string {
	return a
}
func getType(expr ast.Expr) (typeString string, isSlice, isPointer bool, err error) {
	switch expr.(type) {
	case *ast.Ident:
		x := expr.(*ast.Ident)
		return x.Name, false, false, nil
	case *ast.SelectorExpr:
		x := expr.(*ast.SelectorExpr)
		return x.X.(*ast.Ident).Name + "." + x.Sel.Name, false, false, nil
	case *ast.ArrayType:
		tmp := expr.(*ast.ArrayType)
		if tmp.Len != nil {
			tmpLen, ok := tmp.Len.(*ast.BasicLit)
			if !ok {
				return "", false, false, errors.New("array len has unknown type")
			}
			return "[" + tmpLen.Value + "]" + justTypeString(getType(tmp.Elt)), true, false, nil
		}
		return "[]" + justTypeString(getType(tmp.Elt)), true, false, nil
	case *ast.MapType:
		tmp := expr.(*ast.MapType)
		return "map[" + justTypeString(getType(tmp.Key)) + "]" + justTypeString(getType(tmp.Value)), false, false, nil
	case *ast.StarExpr:
		return "*" + justTypeString(getType(expr.(*ast.StarExpr).X)), false, true, nil
	case *ast.FuncType:
		return "", false, false, fmt.Errorf("unsupported type for %#v", expr)
	case *ast.StructType:
		return "", false, false, fmt.Errorf("unsupported type for %#v", expr)
	case *ast.ChanType:
		tmp := expr.(*ast.ChanType)
		switch tmp.Dir {
		case ast.SEND:
			return "chan<- " + justTypeString(getType(tmp.Value)), false, false, nil
		case ast.RECV:
			return "<-chan " + justTypeString(getType(tmp.Value)), false, false, nil
		}
		return "chan " + justTypeString(getType(tmp.Value)), false, false, nil
	case *ast.Ellipsis:
		tmp := expr.(*ast.Ellipsis)
		return "..." + justTypeString(getType(tmp.Elt)), false, false, nil

	}
	return "", false, false, fmt.Errorf("unknown type for %#v", expr)
}

func getImportAndType(fullTypeString string, currentPkg string, currentImportList []*importItem) (importPkg string, builtin bool, typeS string) {
	tmp, _ := strings.CutPrefix(fullTypeString, "*")
	tmp1 := strings.Split(tmp, ".")
	var checkBuiltIn = func(s string) bool {
		return internal.IsInternalType(s)
	}

	switch len(tmp1) {
	case 1: //current package
		return currentPkg, checkBuiltIn(tmp1[0]), tmp1[0]
	case 2: //third package
		pkgName := tmp1[0]
		for _, s := range currentImportList {
			if (s.name == "" || s.name == ".") && strings.HasSuffix(s.importPath, pkgName) {
				return s.importPath, checkBuiltIn(tmp1[1]), tmp1[1]
			} else if s.name != "" && pkgName == s.name {
				return s.importPath, checkBuiltIn(tmp1[1]), tmp1[1]
			}
		}
		return "", true, ""
	default:
		return "", true, ""
	}
}
