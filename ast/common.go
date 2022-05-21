package ast

import (
	"errors"
	"fmt"
	"github.com/linxlib/kapi/doc"
	"github.com/linxlib/kapi/internal"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"runtime"
	"strings"
)

var importFile = make(map[string]string) // 自定义包文件

type ControllerComment struct {
	TagName     string
	Route       string
	TokenHeader string
}

var (
	getAstPackagesCache = make(map[string]*ast.Package)
	parseStructCache    = make(map[string]*doc.StructInfo)
)

// AddImportFile 添加自定义import文件列表
func AddImportFile(k, v string) {
	importFile[k] = v
}

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
			if internal.CheckFileIsExist(filename + "/go.mod") {
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

// EvalSymlinks  Return to relative path . 通过module 游标返回包相对路径
func EvalSymlinks(modPkg, modFile, objPkg string) string {
	if strings.EqualFold(objPkg, "main") { // if main return default path
		return modFile
	}

	if strings.HasPrefix(objPkg, modPkg) {
		return modFile + strings.Replace(objPkg[len(modPkg):], ".", "/", -1)
	}

	// 自定义文件中查找
	tmp := importFile[objPkg]
	if len(tmp) > 0 {
		return tmp
	}

	// get the error space
	panic(fmt.Errorf("can not eval pkg:[%v] must include [%v]", objPkg, modPkg))
}

// GetAstPackage Parsing source file ast structure (with main restriction).解析源文件ast结构(带 main 限制)
func GetAstPackage(objPkg, objFile string) (*ast.Package, bool) {
	key := objPkg + "_" + objFile
	if v, ok := getAstPackagesCache[key]; ok {
		return v, true
	} else {
		fileSet := token.NewFileSet()
		astPkgs, err := parser.ParseDir(fileSet, objFile, func(info os.FileInfo) bool {
			name := info.Name()
			return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
		}, parser.ParseComments)
		if err != nil {
			return nil, false
		}

		// check the package is same.判断 package 是否一致
		for _, pkg := range astPkgs {
			if objPkg == pkg.Name || strings.HasSuffix(objPkg, "/"+pkg.Name) { // find it
				getAstPackagesCache[key] = pkg
				return pkg, true
			}
		}

		// not find . maybe is main package and find main package
		if objPkg == "main" {
			dirs := internal.GetPathDirs(objFile) // get all of dir
			for _, dir := range dirs {
				if !strings.HasPrefix(dir, ".") {
					pkg, b := GetAstPackage(objPkg, objFile+"/"+dir)
					if b {
						getAstPackagesCache[key] = pkg
						return pkg, true
					}
				}
			}
		}
	}

	return nil, false
}

//// AnalysisImport 分析整合import相关信息
//func AnalysisImport(astPkg *ast.Package) map[string]string {
//
//	imports := make(map[string]string)
//	for _, f := range astPkg.Files {
//		for _, p := range f.Imports {
//			k := ""
//			if p.Name != nil {
//				k = p.Name.Name
//			}
//			v := strings.Trim(p.Path.Value, `"`)
//			if len(k) == 0 {
//				n := strings.LastIndex(v, "/")
//				if n > 0 {
//					k = v[n+1:]
//				} else {
//					k = v
//				}
//			}
//			imports[k] = v
//		}
//	}
//
//	return imports
//}

//handleImportSpec 处理导入
func handleImportSpec(p *ast.ImportSpec, imports map[string]string) {
	k := ""
	if p.Name != nil {
		k = p.Name.Name
	}
	v := strings.Trim(p.Path.Value, `"`)
	if len(k) == 0 {
		n := strings.LastIndex(v, "/")
		if n > 0 {
			k = v[n+1:]
		} else {
			k = v
		}
	}
	imports[k] = v
}

//AnalysisControllerFile 处理控制器
func AnalysisControllerFile(astPkg *ast.Package, controllerName string) (imports map[string]string, funcMap map[string]*ast.FuncDecl, comment *ControllerComment) {
	imports = make(map[string]string)
	funcMap = make(map[string]*ast.FuncDecl)
	comment = &ControllerComment{TagName: controllerName}

	for _, f := range astPkg.Files {
		for _, p := range f.Imports {
			handleImportSpec(p, imports)
		}
		for _, d := range f.Decls {
			switch specDecl := d.(type) {
			case *ast.FuncDecl:
				if specDecl.Recv != nil {
					if exp, ok := specDecl.Recv.List[0].Type.(*ast.StarExpr); ok { // Check that the type is correct first before throwing to parser
						if strings.Compare(fmt.Sprint(exp.X), controllerName) == 0 { // is the same struct
							funcMap[specDecl.Name.String()] = specDecl // catch
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
	return
}

// GetImportPkg 分析得出 pkg
func GetImportPkg(i string) string {
	n := strings.LastIndex(i, "/")
	if n > 0 {
		return i[n+1:]
	}
	return i
}
