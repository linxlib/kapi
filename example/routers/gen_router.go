//Package routers code generate by KApi on 2021-10-20 21:12:34. do not edit it.
package routers

import (
	"gitee.com/kirile/kapi"
)

func init() {
	kapi.SetVersion(1634735554)
	kapi.AddGenOne("Hello/List", "/hello2/list", []string{"get"})
	kapi.AddGenOne("Hello/List2", "/hello2/hello/list2", []string{"get"})
}
