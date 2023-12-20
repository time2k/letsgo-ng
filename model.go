package letsgo

/*
* DO NOT MODIFY
 */

//ModelFunc 定义lestgo model func
type ModelFunc func(commp *CommonParams, args ...interface{}) BaseReturnData

//Model letsgo框架加载model方式
func Model(commp *CommonParams, myfunc ModelFunc, args ...interface{}) BaseReturnData {
	defer PanicFunc()

	ret := myfunc(commp, args...)

	return ret
}
