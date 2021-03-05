package letsgo

/*
* DO NOT MODIFY
 */

/*
* debug info define
 */

//DebugInfo debug信息
type DebugInfo struct {
	Info []string `json:"debug"`
}

//Add 向debug中添加信息
func (D *DebugInfo) Add(info string) {
	D.Info = append(D.Info, info)
}

/*
* debug info define end
 */

//ModelFunc 定义lestgo model func
type ModelFunc func(commp CommonParams, args ...interface{}) BaseReturnData

//Model letsgo框架加载model方式
func Model(commp CommonParams, myfunc ModelFunc, args ...interface{}) BaseReturnData {
	defer PanicFunc()

	ret := myfunc(commp, args...)

	return ret
}
