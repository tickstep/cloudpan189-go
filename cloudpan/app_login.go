// 电脑手机客户端API，例如MAC客户端
package cloudpan

import (
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/cloudpan/apiutil"
	"github.com/tickstep/cloudpan189-go/library/crypto"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"github.com/tickstep/cloudpan189-go/library/requester"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type (
	appLoginParams struct {
		CaptchaToken string
		Lt string
		ReturnUrl string
		ParamId string
		ReqId string
		jRsaKey string
	}

	AppLoginResult struct {
		SessionKey string `json:"sessionKey"`
		SessionSecret string `json:"sessionSecret"`
		// 有效期的token
		AccessToken string `json:"accessToken"`
		// token 过期时间点
		AccessTokenExpiresIn int
		RsaPublicKey string
	}

	appSessionResp struct {
		ResCode int `json:"res_code"`
		ResMessage string `json:"res_message"`
		AccessToken string `json:"accessToken"`
		FamilySessionKey string `json:"familySessionKey"`
		FamilySessionSecret string `json:"familySessionSecret"`
		GetFileDiffSpan int `json:"getFileDiffSpan"`
		GetUserInfoSpan int `json:"getUserInfoSpan"`
		IsSaveName string `json:"isSaveName"`
		KeepAlive int `json:"keepAlive"`
		LoginName string `json:"loginName"`
		RefreshToken string `json:"refreshToken"`
		SessionKey string `json:"sessionKey"`
		SessionSecret string `json:"sessionSecret"`
	}

	accessTokenResp struct {
		// token过期时间，默认30天
		ExpiresIn int `json:"expiresIn"`
		AccessToken string `json:"accessToken"`
	}
)

var (
	appClient = requester.NewHTTPClient()
)

func AppLogin(username, password string) (result *AppLoginResult, error *apierror.ApiError) {
	appClient.SetProxy("http://127.0.0.1:8888")
	result = &AppLoginResult{}

	appClient.ResetCookiejar()
	loginParams, err := appGetLoginParams()
	if err != nil {
		logger.Verboseln("get login params error")
		return nil, err
	}
	rsaKey := &strings.Builder{}
	fmt.Fprintf(rsaKey, "-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", loginParams.jRsaKey)
	result.RsaPublicKey = rsaKey.String()
	rsaUserName, _ := crypto.RsaEncrypt([]byte(rsaKey.String()), []byte(username))
	rsaPassword, _ := crypto.RsaEncrypt([]byte(rsaKey.String()), []byte(password))

	urlStr := "https://open.e.189.cn/api/logbox/oauth2/loginSubmit.do"
	headers := map[string]string {
		"Content-Type": "application/x-www-form-urlencoded",
		"Referer": "https://open.e.189.cn/api/logbox/oauth2/unifyAccountLogin.do",
		"Cookie": "LT=" + loginParams.Lt,
		"X-Requested-With": "XMLHttpRequest",
		"REQID": loginParams.ReqId,
		"lt": loginParams.Lt,
	}
	formData := map[string]string {
		"appKey": "8025431004",
		"accountType": "02",
		"userName": "{RSA}" + apiutil.B64toHex(string(crypto.Base64Encode(rsaUserName))),
		"password": "{RSA}" + apiutil.B64toHex(string(crypto.Base64Encode(rsaPassword))),
		"validateCode": "",
		"captchaToken": loginParams.CaptchaToken,
		"returnUrl": loginParams.ReturnUrl,
		"mailSuffix": "@189.cn",
		"dynamicCheck": "FALSE",
		"clientType": "10020",
		"cb_SaveName": "1",
		"isOauth2": "false",
		"state": "",
		"paramId": loginParams.ParamId,
	}

	logger.Verboseln("do request url: " + urlStr)
	body, err1 := appClient.Fetch("POST", urlStr, formData, headers)
	if err1 != nil {
		logger.Verboseln("login redirectURL occurs error: ", err.Error())
		return nil, apierror.NewApiErrorWithError(err)
	}
	logger.Verboseln("response: " + string(body))
	r := &loginResult{}
	if err := json.Unmarshal(body, r); err != nil {
		logger.Verboseln("parse login result json error ", err)
		return nil, apierror.NewFailedApiError(err.Error())
	}
	if r.Result != 0 || r.ToUrl == "" {
		return nil, apierror.NewFailedApiError("登录失败")
	}

	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/getSessionForPC.action?clientType=%s&version=%s&channelId=%s&redirectURL=%s",
		API_URL, "TELEMAC", "1.0.0", "web_cloud.189.cn", url.QueryEscape(r.ToUrl))
	headers = map[string]string {
		"Accept": "application/json;charset=UTF-8",
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err1 = appClient.Fetch("GET", fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("get session info occurs error: ", err.Error())
		return nil, apierror.NewApiErrorWithError(err)
	}
	logger.Verboseln("response: " + string(body))
	rs := &appSessionResp{}
	if err := json.Unmarshal(body, rs); err != nil {
		logger.Verboseln("parse session result json error ", err)
		return nil, apierror.NewFailedApiError(err.Error())
	}
	if rs.ResCode != 0 {
		return nil, apierror.NewFailedApiError("获取session失败")
	}
	result.SessionKey = rs.SessionKey
	result.SessionSecret = rs.SessionSecret

	fullUrl = &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/open/oauth2/getAccessTokenBySsKey.action?sessionKey=%s",
		API_URL, rs.SessionKey)
	timestamp := apiutil.Timestamp()
	signParams := map[string]string {
		"Timestamp": strconv.Itoa(timestamp),
		"sessionKey": rs.SessionKey,
		"AppKey": "601102120",
	}
	headers = map[string]string {
		"AppKey": "601102120",
		"Signature": apiutil.Signature(signParams),
		"Sign-Type": "1",
		"Accept": "application/json",
		"Timestamp": strconv.Itoa(timestamp),
	}
	body, err1 = appClient.Fetch("GET", fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("get accessToken occurs error: ", err.Error())
		return nil, apierror.NewApiErrorWithError(err)
	}
	logger.Verboseln("response: " + string(body))
	atr := &accessTokenResp{}
	if err := json.Unmarshal(body, atr); err != nil {
		logger.Verboseln("parse accessToken result json error ", err)
		return nil, apierror.NewFailedApiError(err.Error())
	}
	result.AccessTokenExpiresIn = atr.ExpiresIn
	result.AccessToken = atr.AccessToken
	return result, nil
}

func appGetLoginParams() (params appLoginParams, error *apierror.ApiError) {
	header := map[string]string {
		"Content-Type": "application/x-www-form-urlencoded",
	}
	fullUrl := &strings.Builder{}
	// use MAC client appid
	fmt.Fprintf(fullUrl, "%s/unifyLoginForPC.action?appId=%s&clientType=%s&returnURL=%s&timeStamp=%d",
		WEB_URL, "8025431004", "10020", "https://m.cloud.189.cn/zhuanti/2020/loginErrorPc/index.html", apiutil.Timestamp())
	logger.Verboseln("do request url: " + fullUrl.String())
	data, err := appClient.Fetch("GET", fullUrl.String(), nil, header)
	if err != nil {
		logger.Verboseln("login redirectURL occurs error: ", err.Error())
		return params, apierror.NewApiErrorWithError(err)
	}
	content := string(data)

	re, _ := regexp.Compile("captchaToken' value='(.+?)'")
	params.CaptchaToken = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("lt = \"(.+?)\"")
	params.Lt = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("returnUrl = '(.+?)'")
	params.ReturnUrl = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("paramId = \"(.+?)\"")
	params.ParamId = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("reqId = \"(.+?)\"")
	params.ReqId = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("j_rsaKey\" value=\"(.+?)\"")
	params.jRsaKey = re.FindStringSubmatch(content)[1]
	return
}