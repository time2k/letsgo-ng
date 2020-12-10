package letsgo

import (
	"strings"
)

//ParamTrim 参数过滤函数
func ParamTrim(param ...string) []string {
	var x []string
	for _, each := range param {
		x = append(x, strings.Trim(each, " \t\n\v\r"))
	}
	return x
}
