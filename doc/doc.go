package doc

import (
	"fmt"
	"regexp"
	"strings"
)

type Model struct {
	Group          string // group 标记
	TagControllers map[string]map[string]DocModel
}

// NewDoc 新建一个doc模板
// 一次Register产生一个
func NewDoc(group string) *Model {
	doc := &Model{Group: group, TagControllers: make(map[string]map[string]DocModel)}
	return doc
}

var pathRouterRegex = regexp.MustCompile(":\\w+")

// replacePathTo 将:id转为{id}这样的格式
func replacePathTo(origin string) string {
	var myorigin = origin
	matches := pathRouterRegex.FindAllString(origin, -1)

	if len(matches) > 0 {
		for _, match := range matches {
			if strings.HasPrefix(match, ":") {
				pathName := strings.TrimPrefix(match, ":")
				pathName = strings.TrimSuffix(pathName, "/")
				pathName = fmt.Sprintf("{%s}", pathName)
				myorigin = strings.ReplaceAll(myorigin, match, pathName)
			}
		}
		return myorigin
	}
	return myorigin
}

// AddOne 添加一个方法
func (m *Model) AddOne(group string, routerPath string, methods []string, summary string, description string, req, resp *StructInfo, tokenHeader string, isDeprecated bool) {
	if m.TagControllers[group] == nil {
		m.TagControllers[group] = make(map[string]DocModel)
	}
	myRouterPath := replacePathTo(routerPath)
	// 解析一个路由方法 并存为文档所需
	m.analysisStructInfo(req)
	m.analysisStructInfo(resp)
	if len(methods) > 0 {
		m.TagControllers[group][methods[0]+" "+myRouterPath] = DocModel{
			RouterPath:   myRouterPath,
			Methods:      methods,
			Summary:      summary,
			Description:  description,
			Req:          req,
			Resp:         resp,
			TokenHeader:  tokenHeader,
			IsDeprecated: isDeprecated,
		}
	}

}
