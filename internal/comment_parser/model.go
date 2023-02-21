package comment_parser

import "strings"

type Comment struct {
	//@CORS
	CORS bool
	//@DEPRECATED
	Deprecated bool //deprecate a method
	HasReq     bool //if a method has request type defined(the second param)
	//@REQ model.RequestType
	RequestType []string // //specify the request type of current method. method's request type has higher priority.
	HasResp     bool     //if a method has result type defined
	//@RESP model.ResponseType
	ResultType []string //specify the result type of current method. method's result type has higher priority.
	//@DESC
	Description []string
	//Name Summary.
	Summary string // if empty, this field will be the Name of it
	//@GET /api/v1/user/list.
	Routes map[string]string // will like map[route]HttpMethod
	//@Anonymous
	Anonymous bool // current method will be anonymous even if `@AUTH` had been set to the controller. not implemented yet.
	//@ROUTE /api/v1.
	Route string // route prefix of current controller
	//@TAG tagname.
	Tag string // will show on Swagger UI as tag
	//@AUTH Authorization.
	AuthorizationHeader string //add a required header parameter to all methods under current controller.
}

func (c *Comment) GetDescription(sep string) string {
	return strings.Join(c.Description, sep)
}
