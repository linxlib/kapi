package model

type MyRequest3 struct {
	A int      `json:"ii" v:"required" default:"1"`
	B []string `query:"bn"`
	C int      `path:"c"`
}

// MyResult3 hhhh
// hsduh
type MyResult3 struct {
	A int `json:"a" v:"required" default:"222"` //aaaa
	B []string
}
