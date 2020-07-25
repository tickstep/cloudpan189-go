// WEB网页端API
package cloudpan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/cloudpan/apiutil"
	"github.com/tickstep/cloudpan189-go/library/crypto"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"github.com/tickstep/cloudpan189-go/library/requester"
	"image/png"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

const (
	// CaptchaName 验证码文件名称
	captchaName = "captcha.png"
)

type (
	loginParams struct {
		CaptchaToken string
		Lt string
		ReturnUrl string
		ParamId string
	}

	loginResult struct {
		Result int `json:"result"`
		Msg string `json:"msg"`
		ToUrl string `json:"toUrl"`
	}

	WebLoginToken struct {
		CookieLoginUser string `json:"cookieLoginUser"`
	}
)

var (
	latestLoginParams loginParams
	client            = requester.NewHTTPClient()
)

func Login(username, password string) (webToken *WebLoginToken, error *apierror.ApiError) {
	client.ResetCookiejar()
	params, err := getLoginParams()
	if err != nil {
		logger.Verboseln("get login params error")
		return nil, err
	}

	err = checkNeedCaptchaCodeOrNot(username, latestLoginParams.Lt)
	if err != nil {
		return nil, err
	}

	// save latest params
	latestLoginParams = params
	token, err := LoginWithCaptcha(username, password, "")
	if err != nil {
		return nil, err
	}
	return token, nil
}

func LoginWithCaptcha(username, password, captchaCode string) (webToken *WebLoginToken, error *apierror.ApiError) {
	//client.ResetCookiejar()
	//latestLoginParams, _ = getLoginParams()

	webToken = &WebLoginToken{}
	if latestLoginParams.CaptchaToken == "" {
		latestLoginParams, _ = getLoginParams()
	}

	r, err := doLoginAct(username, password, captchaCode, latestLoginParams.CaptchaToken,
		latestLoginParams.ReturnUrl, latestLoginParams.ParamId, latestLoginParams.Lt)
	if err != nil || r.Msg != "登录成功" {
		logger.Verboseln("login failed ", err)
		return webToken, apierror.NewFailedApiError(err.Error())
	}
	// request toUrl to get COOKIE_LOGIN_USER cookie
	header := map[string]string {
		"lt":           latestLoginParams.Lt,
		"Content-Type": "application/x-www-form-urlencoded",
		"Referer":      "https://open.e.189.cn/",
	}
	client.Fetch("GET", r.ToUrl, nil, header)

	cloudpanUrl := &url.URL{
		Scheme: "http",
		Host:   "cloud.189.cn",
		Path: "/",
	}
	cks := client.Jar.Cookies(cloudpanUrl)
	for _, cookie := range cks {
		if cookie.Name == "COOKIE_LOGIN_USER" {
			webToken.CookieLoginUser = cookie.Value
			break
		}
	}

	return
}

func GetCaptchaImage() (savePath string, error *apierror.ApiError) {
	if latestLoginParams.CaptchaToken == "" {
		latestLoginParams, _ = getLoginParams()
	}

	removeCaptchaPath()
	picUrl := AUTH_URL + "/picCaptcha.do?token=" + latestLoginParams.CaptchaToken
	// save img to file
	return saveCaptchaImg(picUrl)
}

func getLoginParams() (params loginParams, error *apierror.ApiError) {
	header := map[string]string {
		"Content-Type": "application/x-www-form-urlencoded",
	}
	data, err := client.Fetch("GET", WEB_URL+ "/udb/udb_login.jsp?pageId=1&redirectURL=/main.action",
		nil, header)
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
	return
}

func checkNeedCaptchaCodeOrNot(username, lt string) (error *apierror.ApiError) {
	url := AUTH_URL + "/needcaptcha.do"
	rsa, err := crypto.RsaEncrypt([]byte(apiutil.RsaPublicKey), []byte(username))
	if err != nil {
		return apierror.NewApiErrorWithError(err)
	}
	postData := map[string]string {
		"accountType": "01",
		"userName": "{RSA}" + apiutil.B64toHex(string(crypto.Base64Encode(rsa))),
		"appKey": "cloud",
	}
	header := map[string]string {
		"lt": lt,
		"Content-Type": "application/x-www-form-urlencoded",
		"Referer": "https://open.e.189.cn/",
	}
	body, err := client.Fetch("POST", url, postData, header)
	if err != nil {
		logger.Verboseln("get captcha code error: ", err.Error())
		return apierror.NewApiErrorWithError(err)
	}
	text := string(body)
	if text != "0" {
		// need captcha
		return apierror.NewApiError(apierror.ApiCodeNeedCaptchaCode, "需要验证码")
	}
	return
}

func saveCaptchaImg(imgURL string) (savePath string, error *apierror.ApiError) {
	logger.Verboseln("try to download captcha image: ", imgURL)
	imgContents, err := client.Fetch("GET", imgURL, nil, nil)
	if err != nil {
		return "", apierror.NewApiErrorWithError(fmt.Errorf("获取验证码失败, 错误: %s", err))
	}

	_, err = png.Decode(bytes.NewReader(imgContents))
	if err != nil {
		return "", apierror.NewApiErrorWithError(fmt.Errorf("验证码解析错误: %s", err))
	}

	savePath = captchaPath()
	return savePath, apierror.NewApiErrorWithError(ioutil.WriteFile(savePath, imgContents, 0777))
}

func captchaPath() string {
	return filepath.Join(os.TempDir(), captchaName)
}

func removeCaptchaPath() error {
	return os.Remove(captchaPath())
}

func doLoginAct(username, password, validateCode, captchaToken, returnUrl, paramId, lt string) (result *loginResult, error *apierror.ApiError) {
	url := AUTH_URL + "/loginSubmit.do"
	rsaUserName, _ := crypto.RsaEncrypt([]byte(apiutil.RsaPublicKey), []byte(username))
	rsaPassword, _ := crypto.RsaEncrypt([]byte(apiutil.RsaPublicKey), []byte(password))
	data := map[string]string {
		"appKey": "cloud",
		"accountType": "01",
		"userName": "{RSA}" + apiutil.B64toHex(string(crypto.Base64Encode(rsaUserName))),
		"password": "{RSA}" + apiutil.B64toHex(string(crypto.Base64Encode(rsaPassword))),
		"validateCode": validateCode,
		"captchaToken": captchaToken,
		"returnUrl": returnUrl,
		"mailSuffix": "@189.cn",
		"paramId": paramId,
	}
	header := map[string]string {
		"lt": lt,
		"Content-Type": "application/x-www-form-urlencoded",
		"Referer": "https://open.e.189.cn/",
	}

	body, err := client.Fetch("POST", url, data, header)
	if err != nil {
		logger.Verboseln("login with captch error ", err)
		return nil, apierror.NewFailedApiError(err.Error())
	}

	r := &loginResult{}
	if err := json.Unmarshal(body, r); err != nil {
		logger.Verboseln("parse login resutl json error ", err)
		return nil, apierror.NewFailedApiError(err.Error())
	}
	return r, nil
}