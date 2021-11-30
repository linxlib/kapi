package model

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name,omitempty"`
	PID  int64  `json:"pid"`
	PPID int64  `json:"ppid"`
}
