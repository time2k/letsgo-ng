package letsgo

//CommonHeader 通用头
type CommonHeader struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	//ErrCode int
}

//CommonResp 通用返回
type CommonResp struct {
	CommonHeader `json:"header"`
	Body         interface{} `json:"body"`
}

//CommonRespWithDebug 带有debug信息的返回
type CommonRespWithDebug struct {
	CommonResp
	DebugInfo
}

//BaseReturnData letsgo model 默认数据返回封装，你可以根据这个重新在此文件定义自定义封装
type BaseReturnData struct {
	Status int
	Msg    string
	Body   interface{}
	CommonParams
}

//CommonRespNew 通用返回
type CommonRespNew struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Debug   []string    `json:"debug,omitempty"`
}

//FormatNew 格式化方法
func (BD *BaseReturnData) FormatNew() interface{} {
	//init
	ret := CommonRespNew{}

	ret.Code = BD.Status
	ret.Message = BD.Msg
	ret.Data = BD.Body
	if BD.CommonParams.GetParam("debug") != "" && len(BD.CommonParams.Debug.Info) != 0 {
		ret.Debug = BD.CommonParams.Debug.Info
	}

	return ret
}
