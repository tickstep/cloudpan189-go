package apierror

// ErrorResp 默认的错误信息
type ErrorResp struct {
	ErrorCode string `json:"errorCode"`
	ErrorMsg string `json:"errorMsg"`
}

type SuccessResp struct {
	// Success 是否成功。true为成功，false或者没有返回则为失败
	Success bool `json:"success"`
}