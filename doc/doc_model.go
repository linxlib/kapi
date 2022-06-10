package doc

// DocModel Model
type DocModel struct {
	RouterPath   string
	Methods      []string
	Summary      string
	Description  string
	Req, Resp    *StructInfo
	TokenHeader  string
	IsDeprecated bool
}
