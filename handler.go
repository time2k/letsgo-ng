package letsgo

/*
* DO NOT MODIFY
 */

import (
	"letsgo/libs"

	"github.com/labstack/echo"
)

//HandlerFunc 定义lestgo handler func
type HandlerFunc func(echo.Context) error

//Handler letsgo框架标准handler
func Handler(myfunc HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer libs.PanicFunc()
		return myfunc(c)
	}
}
