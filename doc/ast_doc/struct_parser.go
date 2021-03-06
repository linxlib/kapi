package ast_doc

import (
	"fmt"
	"github.com/linxlib/kapi/doc"
	"github.com/linxlib/kapi/internal"
	"go/ast"
	"go/token"
	"reflect"
	"strings"
)

type structParser struct {
	ModPkg, ModFile string
}

// NewStructAnalysis 新建一个导出结构体类
func NewStructAnalysis(modPkg, modFile string) *structParser {
	result := &structParser{ModPkg: modPkg, ModFile: modFile}
	return result
}

// ParseStruct 解析结构体定义及相关信息
func (a *structParser) ParseStruct(astPkg *ast.Package, structName string, isarray bool) *doc.StructInfo {
	if astPkg == nil {
		return nil
	}
	//internal.Log.Debugf("ParseStruct:%s from:%s",structName,astPkg.Name)
	if internal.IsInternalType(structName) { // 内部类型
		return &doc.StructInfo{
			Name:    structName,
			IsArray: isarray,
		}
	}

	key := fmt.Sprintf("%s_%s_%t", astPkg.Name, structName, isarray)
	if v, ok := parseStructCache[key]; ok {
		return v
	} else {
		for _, fl := range astPkg.Files {
			for _, d := range fl.Decls {
				switch specDecl := d.(type) {
				case *ast.GenDecl:
					for _, subitem := range specDecl.Specs {
						switch specDecl.Tok {
						case token.TYPE:
							spec := subitem.(*ast.TypeSpec)
							switch st := spec.Type.(type) {
							case *ast.StructType:
								if spec.Name.Name == structName { // find it
									info := &doc.StructInfo{
										Items:   a.structFieldInfo(astPkg, st),
										Note:    "",
										Name:    structName,
										Pkg:     astPkg.Name,
										IsArray: isarray,
									}
									if specDecl.Doc != nil { // 如果有注释
										for _, v := range specDecl.Doc.List { // 结构体注释
											t := strings.TrimSpace(strings.TrimPrefix(v.Text, "//"))
											if strings.HasPrefix(t, structName) { // find note
												t = strings.TrimSpace(strings.TrimPrefix(t, structName))
												info.Note += t
											}
										}
									}
									parseStructCache[key] = info
									return info
								}

							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (a *structParser) structFieldInfo(astPkg *ast.Package, structType *ast.StructType) (items []*doc.ElementInfo) {
	if structType == nil || structType.Fields == nil {
		return
	}

	items = make([]*doc.ElementInfo, 0)
	importMP := make(map[string]string)
	for _, f := range astPkg.Files {
		for _, p := range f.Imports {
			handleImportSpec(p, importMP)
		}
	}

	//遍历结构体字段
	for _, field := range structType.Fields.List {
		info := &doc.ElementInfo{}

		for _, fieldName := range field.Names {
			info.Name += fieldName.Name
		}
		if info.Name == "" { //处理没有字段名的 （类似继承）
			if exp, ok := field.Type.(*ast.Ident); ok { //比如报名,函数名,变量名
				a.handleIdent(astPkg, exp, info)
			} else if exp, ok := field.Type.(*ast.StarExpr); ok {
				switch x := exp.X.(type) {
				case *ast.SelectorExpr: // 选择结构,类似于a.b的结构
					a.handleSelectorExpr(x, info, importMP)
				case *ast.Ident: //报名,函数名,变量名
					a.handleIdent(astPkg, x, info)
				}
			} else if exp, ok := field.Type.(*ast.SelectorExpr); ok {
				a.handleSelectorExpr(exp, info, importMP)
			}
			items = append(items, info.TypeRef.Items...)

			continue
		}

		// 判断是否是导出属性(导出属性才允许)(首字母大写)
		strArray := []rune(info.Name)
		if len(strArray) > 0 && (strArray[0] >= 97 && strArray[0] <= 122) { // 首字母小写
			continue
		}

		if field.Tag != nil {
			info.Tag = strings.Trim(field.Tag.Value, "`")
			tag := reflect.StructTag(info.Tag)
			//TODO: 其他标签也处理忽略逻辑
			tagStr := tag.Get("json")
			if tagStr == "-" { // 忽略的json字段
				continue
			}
		}
		if field.Comment != nil {
			info.Note = strings.TrimSpace(field.Comment.Text())
		}
		if field.Doc != nil {
			info.Note += strings.TrimSpace(field.Doc.List[0].Text)
		}

		switch exp := field.Type.(type) {
		case *ast.SelectorExpr: // 非本文件包
			a.handleSelectorExpr(exp, info, importMP)
		case *ast.ArrayType: //数组
			info.IsArray = true
			switch x := exp.Elt.(type) {
			case *ast.SelectorExpr: // 非本文件包
				a.handleSelectorExpr(x, info, importMP)
			case *ast.Ident:
				a.handleIdent(astPkg, x, info)
			case *ast.StarExpr:
				switch x1 := x.X.(type) {
				case *ast.SelectorExpr: // 非本文件包
					a.handleSelectorExpr(x1, info, importMP)
				case *ast.Ident:
					a.handleIdent(astPkg, x1, info)
				}
			case *ast.ArrayType: //这里支持二维数组
				info.IsTDArray = true
				switch y := x.Elt.(type) {
				case *ast.SelectorExpr: // 非本文件包
					a.handleSelectorExpr(y, info, importMP)
				case *ast.Ident:
					a.handleIdent(astPkg, y, info)
				case *ast.StarExpr:
					switch x1 := y.X.(type) {
					case *ast.SelectorExpr: // 非本文件包
						a.handleSelectorExpr(x1, info, importMP)
					case *ast.Ident:
						a.handleIdent(astPkg, x1, info)
					}
				}

			}
		case *ast.StarExpr: //类型
			switch x := exp.X.(type) {
			case *ast.SelectorExpr: // 非本文件包
				a.handleSelectorExpr(x, info, importMP)
			case *ast.Ident:
				a.handleIdent(astPkg, x, info)
			}
		case *ast.Ident: // 本文件
			a.handleIdent(astPkg, exp, info)
		case *ast.MapType: // map
			key := ""
			value := ""
			switch x := exp.Key.(type) {
			case *ast.Ident:
				key = x.Name
			case *ast.StarExpr:
				switch x1 := x.X.(type) {
				case *ast.SelectorExpr: // 非本文件包
					key = x1.Sel.Name
				case *ast.Ident:
					key = x1.Name
				}
			case *ast.SelectorExpr: // 非本文件包
				key = x.Sel.Name
			}
			switch x := exp.Value.(type) {
			case *ast.Ident:
				value = x.Name
			case *ast.StarExpr:
				switch x1 := x.X.(type) {
				case *ast.SelectorExpr: // 非本文件包
					value = x1.Sel.Name
				case *ast.Ident:
					value = x1.Name
				}
			case *ast.SelectorExpr: // 非本文件包
				value = x.Sel.Name
			}
			info.Type = fmt.Sprintf("map (%v,%v)", key, value)
		}

		if len(info.Type) == 0 {
			panic(fmt.Sprintf("不支持的类型 : %v", field.Type))
		}

		items = append(items, info)
	}
	return items
}

func (a *structParser) handleSelectorExpr(exp *ast.SelectorExpr, info *doc.ElementInfo, importMP map[string]string) { // 非本文件包
	info.Type = exp.Sel.Name
	if !internal.IsInternalType(info.Type) { // 非基础类型(time)
		if x, ok := exp.X.(*ast.Ident); ok {
			if v, ok := importMP[x.Name]; ok {
				objFile := EvalSymlinks(a.ModPkg, a.ModFile, v)
				objPkg := GetImportPkg(v)
				astFile, _b := GetAstPackage(objPkg, objFile)
				if _b {
					info.TypeRef = a.ParseStruct(astFile, info.Type, false)
				}
			}
		}
	}
}

// handleIdent 处理类型
func (a *structParser) handleIdent(astPkg *ast.Package, exp *ast.Ident, info *doc.ElementInfo) { // 本文件
	info.Type = exp.Name
	if !internal.IsInternalType(info.Type) { // 非基础类型
		info.TypeRef = a.ParseStruct(astPkg, info.Type, false)
	}
}
