package model

type User struct {
	ID       int64  `json:"-"`
	Name     string `json:"name"`
	Gender   int    `json:"gender"`
	NickName string `json:"nick_name"`
	Phone    string `json:"phone"`
}
