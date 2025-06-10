package letsgo

/*
* DO NOT MODIFY
 */

import (
	"github.com/labstack/echo/v4"
)

// HandlerFunc 定义lestgo handler func
type HandlerFunc func(commp *CommonParams) error

// Handler letsgo框架加载handler
func Handler(myfunc HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer PanicFunc()

		//params check
		//通用参数处理，通用参数包括letsgo框架指针通过此结构体传递到model
		return myfunc(GenCommparams(c))
	}
}

// GenCommparams 通用参数处理
func GenCommparams(c echo.Context) *CommonParams {
	commp := CommonParams{}
	commp.Init()

	if c != nil {
		commp.HTTPContext = c

		//通用参数处理
		params := ParamTrim(c.QueryParam("pcode"), c.QueryParam("version"), c.QueryParam("lon"), c.QueryParam("lat"), c.QueryParam("ip"), c.QueryParam("did"), c.QueryParam("_debug"))
		commp.SetParam("pcode", params[0])
		commp.SetParam("version", params[1])
		commp.SetParam("lon", params[2])
		commp.SetParam("lat", params[3])
		if params[4] != "" {
			commp.SetParam("ip", params[4])
		} else {
			commp.SetParam("ip", c.RealIP())
		}
		commp.SetParam("did", params[5])
		commp.SetParam("useragent", c.Request().UserAgent())
		commp.SetParam("debug", params[6])
	}

	return &commp
}
