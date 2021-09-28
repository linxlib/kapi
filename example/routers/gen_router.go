//Package routers code generate by KApi on 2021-09-12 17:55:34. do not edit it.
package routers

import (
	"gitee.com/kirile/kapi"
)

func init() {
	kapi.SetVersion(1631440534)
	kapi.AddGenOne("Hello/List", "/hello2/list", []string{"get"})
	kapi.AddGenOne("Hello/List2", "/hello2/hello/list2", []string{"get"})
}
