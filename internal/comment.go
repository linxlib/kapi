package internal

import (
	"regexp"
)

// GetCommentAfterPrefixRegex 解析注释 分离前缀和注释内容
func GetCommentAfterPrefixRegex(fullComment string, name string) (prefix string, comment string, b bool) {
	var myRegex = regexp.MustCompile(`(@\w+)\s*(\S*)`)
	var myRegex1 = regexp.MustCompile(`(\/\/\s*` + name + `)\s*(\S*)`)
	//tmp := strings.TrimSpace(strings.TrimPrefix(fullComment, "//")) //@TAG content...

	matches1 := myRegex1.FindStringSubmatch(fullComment)
	if len(matches1) == 3 {
		prefix = matches1[1]
		comment = matches1[2]
		b = true
		return
	} else if len(matches1) == 2 {
		prefix = matches1[1]
		comment = ""
		b = true
		return
	} else {
		matches := myRegex.FindStringSubmatch(fullComment)
		if len(matches) == 3 {
			prefix = matches[1]
			comment = matches[2]
			b = true
			return
		} else if len(matches) == 2 {
			prefix = matches[1]
			comment = ""
			b = true
			return
		} else {
			return "", "", false
		}
	}

}
