package mindoc

type Root struct {
	Title   string   //服务标题
	BaseURL string   //基本url
	BaseApi string   //基本path
	Scheme  string   //http(s)
	Descs   []string //描述
	Apis    []Api    //apis
	Footers []string //附
}

type Param struct {
	Name          string
	Optional      bool
	OptionalText  string
	Default       string
	ParamType     string
	ParamLocation string
	Desc          string
}

type Resp struct {
	Name      string
	ParamType string
	Desc      string
}

type Api struct {
	Index               int
	Name                string
	Desc                string
	Method              string
	RoutePath           string
	Params              []Param
	Resps               []Resp
	BodyExample         string
	SuccessExampleJson  string
	FailedExampleJson   string
	RequestExampleCode  string
	ResponseExampleCode string
}
