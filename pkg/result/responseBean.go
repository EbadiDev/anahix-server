package result

type ResponseSuccessBean struct {
	Code uint32      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(data interface{}) *ResponseSuccessBean {
	return &ResponseSuccessBean{
		Code: 200,
		Msg:  "success",
		Data: data,
	}
}

type ResponseErrorBean struct {
	Code uint32 `json:"code"`
	Msg  string `json:"msg"`
}

func Error(errCode uint32, errMsg string) *ResponseErrorBean {
	return &ResponseErrorBean{
		Code: errCode,
		Msg:  errMsg,
	}
}
