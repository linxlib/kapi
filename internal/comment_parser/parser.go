package comment_parser

import (
	"regexp"
	"strings"
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
	comments []string
	name     string
	logger   ParserLogger
}

func NewParser(name string, comments []string) *Parser {
	return &Parser{name: name, comments: comments, logger: &_logger{}}
}

func (p *Parser) Parse(groupRoute string) *Comment {
	mc := &Comment{
		Deprecated:  false,
		RequestType: []string{},
		ResultType:  []string{},
		Description: []string{},
		Route:       "",
		Summary:     p.name, //default is method name
		Routes:      make(map[string]string),
		Anonymous:   false,
		Tag:         "",
	}
	for _, comment := range p.comments {
		prefix, comment := getCommentAfterPrefixRegex(comment, p.name)
		switch prefix {
		case "@DEPRECATED":
			mc.Deprecated = true
		case "@ROUTE":
			mc.Route = comment
		case "@TAG":
			mc.Tag = comment
		case "@AUTH":
			if comment == "" {
				comment = "Authorization"
			}
			mc.AuthorizationHeader = comment
		case "@REQ":
			mc.HasReq = true
			mc.RequestType = strings.Split(comment, ".")
		case "@RESP":
			mc.HasResp = true
			mc.ResultType = strings.Split(comment, ".")
		case "@DESC":

			mc.Description = append(mc.Description, comment) //we can have multiple @DESC to multiline description
		case "@GET", "@POST", "@PUT", "@DELETE", "@PATCH", "@OPTION", "@HEAD":
			httpMethod := strings.ToUpper(strings.TrimPrefix(prefix, "@"))
			routerPath := comment

			if groupRoute != "" {
				routerPath = groupRoute + routerPath
				routerPath = strings.TrimSuffix(routerPath, "/")
			}
			if routerPath == "/" || strings.TrimSpace(routerPath) == "" {
				break
			}
			mc.Routes[routerPath] = httpMethod
		case p.name: //if prefix is equal to method name
			mc.Summary = comment // summary can have only one
		default: //not defined comments
			mc.Description = append(mc.Description, comment)
		}

	}
	return mc
}

// getCommentAfterPrefixRegex 解析注释 分离前缀和注释内容
func getCommentAfterPrefixRegex(lineComment string, name string) (prefix string, comment string) {
	var myRegex = regexp.MustCompile(`\s*(` + name + `|@\w+)\s*(.*)|(.*)`)

	matches := myRegex.FindStringSubmatch(lineComment)
	if len(matches) == 4 && matches[1] != "" && matches[2] != "" { //like  `//name xxx` or `//@GET xxxx`
		prefix = matches[1]
		comment = matches[2]
		return
	} else if len(matches) == 4 && matches[1] != "" { //like  `//name or //@GET`
		prefix = matches[1]
		comment = ""
		return
	} else { // common comment
		return "", matches[3]
	}
}
