package model

type Company struct {
	ID       int64  `json:"id"`        //主键
	Name     string `json:"name"`      //单位名称
	PID      int64  `json:"pid"`       //父id
	Level    int    `json:"level"`     //层
	Sort     int    `json:"sort"`      //排序
	InDate   int64  `json:"in_date"`   //插入时间
	EditDate int64  `json:"edit_date"` //修改时间
	DelDate  int64  `json:"del_date"`  //删除时间
}
