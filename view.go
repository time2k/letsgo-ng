package letsgo

//BaseReturnData letsgo model 默认数据返回封装，你可以根据这个重新在此文件定义自定义封装
type BaseReturnData struct {
	Status int
	Msg    string
	Body   interface{}
}

//CommonRespNew 通用返回
type CommonRespNew struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Debug   []string    `json:"debug,omitempty"`
}

//FormatNew 格式化方法
func (BD *BaseReturnData) FormatNew(commp *CommonParams) interface{} {
	//init
	ret := CommonRespNew{}

	ret.Code = BD.Status
	ret.Message = BD.Msg
	ret.Data = BD.Body
	if commp.GetParam("debug") != "" && len(commp.Debug.Info) != 0 {
		ret.Debug = commp.Debug.Info
	}

	return ret
}
