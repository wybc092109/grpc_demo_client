package response

type ResponseBean struct {
	Status uint32      `json:"status"`
	Info   string      `json:"info"`
	Data   interface{} `json:"data"`
}

func Success(data interface{}) *ResponseBean {
	return &ResponseBean{0, "OK", data}
}

func Error(errCode uint32, errInfo string) *ResponseBean {
	return &ResponseBean{errCode, errInfo, struct{}{}}
}
