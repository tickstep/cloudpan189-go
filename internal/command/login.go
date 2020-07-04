package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdliner"
	_ "github.com/tickstep/cloudpan189-go/library/requester"
)

func RunLogin(username, password string) (cookieLoginUser string, error error) {
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

	// try login directly
	cookieLoginUser, apiErr := cloudpan.Login(username, password)
	if apiErr != nil {
		if apiErr.Code == apierror.ApiCodeNeedCaptchaCode {
			for i := 0; i < 10; i++ {
				// 需要认证码
				savePath, apiErr := cloudpan.GetCaptchaImage()
				if apiErr != nil {
					fmt.Errorf("获取认证码错误")
					return cookieLoginUser, apiErr
				}
				fmt.Printf("打开以下路径, 以查看验证码\n%s\n\n", savePath)
				vcode, err := line.State.Prompt("请输入验证码 > ")
				if err != nil {
					return cookieLoginUser, err
				}
				cookieLoginUser, apiErr = cloudpan.LoginWithCaptcha(username, password, vcode)
				if apiErr != nil {
					return "", apiErr
				} else {
					return
				}
			}

		} else {
			return "", fmt.Errorf("登录失败")
		}
	}
	return
}
