package apierror

// ErrorResp 默认的错误信息
type ErrorResp struct {
	ErrorCode string `json:"errorCode"`
	ErrorMsg string `json:"errorMsg"`
}
