package binding

import (
	"net/http"
)

var Query = queryBinding{}

type queryBinding struct{}

func (queryBinding) Name() string {
	return "query"
}

func (queryBinding) Bind(req *http.Request, obj interface{}) error {
	values := req.URL.Query()
	if err := mapQuery(obj, values); err != nil {
		return err
	}
	return validate(obj)
}
func mapQuery(ptr interface{}, form map[string][]string) error {
	return mapFormByTag(ptr, form, "query")
}
