package ast_doc

import (
	"errors"
	"fmt"
	"github.com/linxlib/kapi/internal"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"runtime"
	"strings"
)

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

// GetImportPkg 分析得出 pkg
func GetImportPkg(i string) string {
	n := strings.LastIndex(i, "/")
	if n > 0 {
		return i[n+1:]
	}
	return i
}
