package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cmder/cmdliner"
	_ "github.com/tickstep/cloudpan189-go/library/requester"
)

func RunLogin(username, password string) (err error) {
	line := cmdliner.NewLiner()
	defer line.Close()

	if username == "" {
		username, err = line.State.Prompt("请输入用户名(手机号/邮箱/用户名), 回车键提交 > ")
		if err != nil {
			return
		}
	}

	if password == "" {
		// liner 的 PasswordPrompt 不安全, 拆行之后密码就会显示出来了
		fmt.Printf("请输入密码(输入的密码无回显, 确认输入完成, 回车提交即可) > ")
		password, err = line.State.PasswordPrompt("")
		if err != nil {
			return
		}
	}

	return
}
