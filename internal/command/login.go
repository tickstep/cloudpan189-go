package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/cmder/cmdliner"
	"github.com/tickstep/cloudpan189-go/library/logger"
	_ "github.com/tickstep/cloudpan189-go/library/requester"
)

func RunLogin(username, password string) (usernameStr, passwordStr string, webToken cloudpan.WebLoginToken, appToken cloudpan.AppLoginToken, error error) {
	line := cmdliner.NewLiner()
	defer line.Close()

	if username == "" {
		username, error = line.State.Prompt("请输入用户名(手机号/邮箱/别名), 回车键提交 > ")
		if error != nil {
			return
		}
	}

	if password == "" {
		// liner 的 PasswordPrompt 不安全, 拆行之后密码就会显示出来了
		fmt.Printf("请输入密码(输入的密码无回显, 确认输入完成, 回车提交即可) > ")
		password, error = line.State.PasswordPrompt("")
		if error != nil {
			return
		}
	}

	// app login
	atoken, apperr := cloudpan.AppLogin(username, password)
	if apperr != nil {
		fmt.Println("APP登录失败：", apperr)
		return "", "", webToken, appToken, fmt.Errorf("登录失败")
	}

	// web cookie
	wtoken := &cloudpan.WebLoginToken{}
	cookieLoginUser := cloudpan.RefreshCookieToken(atoken.SessionKey)
	if cookieLoginUser != "" {
		logger.Verboseln("get COOKIE_LOGIN_USER by session key")
		wtoken.CookieLoginUser = cookieLoginUser
	} else {
		// try login directly
		wtoken, apperr = cloudpan.Login(username, password)
		if apperr != nil {
			if apperr.Code == apierror.ApiCodeNeedCaptchaCode {
				for i := 0; i < 10; i++ {
					// 需要认证码
					savePath, apiErr := cloudpan.GetCaptchaImage()
					if apiErr != nil {
						fmt.Errorf("获取认证码错误")
						return "", "", webToken, appToken, apiErr
					}
					fmt.Printf("打开以下路径, 以查看验证码\n%s\n\n", savePath)
					vcode, err := line.State.Prompt("请输入验证码 > ")
					if err != nil {
						return "", "", webToken, appToken, err
					}
					wtoken, apiErr = cloudpan.LoginWithCaptcha(username, password, vcode)
					if apiErr != nil {
						return "", "", webToken, appToken, apiErr
					} else {
						return
					}
				}

			} else {
				return "", "", webToken, appToken, fmt.Errorf("登录失败")
			}
		}
	}

	webToken = *wtoken
	appToken = *atoken
	usernameStr = username
	passwordStr = password
	return
}
