package Libs

import (
	"strings"
)

//Param_trim 参数过滤函数
func Param_trim(param... string) ([]string) {
	var x []string
	for _,each := range param {
		x = append(x, strings.Trim(each, " \t\n\v\r"))
	}
	return x
}