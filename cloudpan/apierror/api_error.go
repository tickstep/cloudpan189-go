package apierror

const (
	// 成功
	ApiCodeOk ApiCode = 0
	// 成功
	ApiCodeNeedCaptchaCode ApiCode = 10
	// 会话/Token已过期
	ApiCodeTokenExpiredCode ApiCode = 11
	// 文件不存在
	ApiCodeFileNotFoundCode ApiCode = 12
	// 上传文件失败
	ApiCodeUploadFileStatusVerifyFailed = 13
	// 上传文件数据偏移值校验失败
	ApiCodeUploadOffsetVerifyFailed = 14
	// 服务器上传文件不存在
	ApiCodeUploadFileNotFound = 15
	// 失败
	ApiCodeFailed ApiCode = 999
)

type ApiCode int

type ApiError struct {
	Code ApiCode
	Err string
}

func NewApiError(code ApiCode, err string) *ApiError {
	return &ApiError {
		code,
		err,
	}
}

func NewApiErrorWithError(err error) *ApiError {
	if err == nil {
		return NewApiError(ApiCodeOk, "")
	} else {
		return NewApiError(ApiCodeFailed, err.Error())
	}
}

func NewOkApiError() *ApiError {
	return NewApiError(ApiCodeOk, "")
}

func NewFailedApiError(err string) *ApiError {
	return NewApiError(ApiCodeFailed, err)
}

func (a *ApiError) SetErr(code ApiCode, err string) {
	a.Code = code
	a.Err = err
}

func (a *ApiError) Error() string {
	return a.Err
}

func (a *ApiError) ErrCode() ApiCode {
	return a.Code
}
