package binding

import (
	"github.com/gin-gonic/gin"
)

var (
	Path = pathBinding{}
)

type pathBinding struct{}

func (pathBinding) Name() string {
	return "path"
}

func (b pathBinding) Bind(c *gin.Context, obj interface{}) error {
	m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	return b.BindUri(m, obj)
}

func (pathBinding) BindUri(m map[string][]string, obj interface{}) error {
	if err := mapUri(obj, m); err != nil {
		return err
	}
	return validate(obj)
}
func mapUri(ptr interface{}, m map[string][]string) error {
	return mapFormByTag(ptr, m, "path")
}
