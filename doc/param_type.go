package doc

type ParamType int

const (
	ParamTypeQuery ParamType = iota
	ParamTypeHeader
	ParamTypeForm
	ParamTypePath
)
