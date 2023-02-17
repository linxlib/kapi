package comment_parser

type Comment struct {
	Deprecated          bool //@DEPRECATED
	HasReq              bool
	RequestType         []string //@REQ model.RequestType
	HasResp             bool
	ResultType          []string          //@RESP model.ResponseType
	Description         []string          //@DESC
	Summary             string            //Name ...
	Routes              map[string]string //@GET /api/v1/user/list
	Anonymous           bool              //@Anonymous
	Route               string            //@ROUTE
	Tag                 string            //@TAG tagname
	AuthorizationHeader string            //@AUTH

}
