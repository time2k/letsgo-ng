package Libs

//BaseReturnData letsgo model 默认数据返回封装，你可以根据这个重新在此文件定义自定义封装
type BaseReturnData struct {
	Status    int
	Msg       string
	Body      interface{}
	IsDebug   bool
	DebugInfo []DebugModel
}

//Format 格式化方法
func (BD *BaseReturnData) Format() interface{} {
	//init
	ret := CommonResp{}

	ret.CommonHeader.Status = BD.Status
	ret.CommonHeader.Msg = BD.Msg
	ret.Body = BD.Body
	if BD.IsDebug == true {
		ret_debug := CommonRespWithDebug{}
		ret_debug.CommonResp = ret

		for _, ec := range BD.DebugInfo {
			for _, ec2 := range ec.Info {
				ret_debug.DebugModel.Add(ec2)
				ret_debug.DebugModel.Add("----")
			}
		}
		return ret_debug
	}

	return ret
}

//CommonRespNew 通用返回
type CommonRespNew struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

//CommonRespWithDebugNew 带有debug信息的返回
type CommonRespWithDebugNew struct {
	CommonRespNew
	DebugModel
}

//FormatNew 格式化方法
func (BD *BaseReturnData) FormatNew() interface{} {
	//init
	ret := CommonRespNew{}

	ret.Code = BD.Status
	ret.Message = BD.Msg
	ret.Data = BD.Body
	if BD.IsDebug == true {
		ret_debug := CommonRespWithDebugNew{}
		ret_debug.CommonRespNew = ret

		for _, ec := range BD.DebugInfo {
			for _, ec2 := range ec.Info {
				ret_debug.DebugModel.Add(ec2)
				ret_debug.DebugModel.Add("----")
			}
		}
		return ret_debug
	}

	return ret
}
