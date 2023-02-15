package utils

import (
	"strings"
	"text/template"
)

var T = template.New("").Funcs(template.FuncMap{
	"ToLower": func(ori string) string {
		return strings.ToLower(ori)
	},
})
