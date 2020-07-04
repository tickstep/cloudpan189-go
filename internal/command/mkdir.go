package command

import (
	"fmt"
	"path"
	"strings"
)

func RunMkdir(name string) {
	if path.IsAbs(name) || strings.ContainsAny(name, "/") {
		fmt.Println("只支持在工作目录下创建文件夹")
		return
	}
	if name == "" {
		fmt.Println("请输入文件夹名")
		return
	}
	name = strings.TrimSpace(name)
	if !CheckFileNameValid(name) {
		fmt.Println("文件夹名不能包含特殊字符：" + FileNameSpecialChars)
		return
	}
	activeUser := GetActiveUser()
	r, err := GetActivePanClient().Mkdir(activeUser.WorkdirFileEntity.FileId, name)
	if err != nil {
		fmt.Printf("创建目录 %s 失败, %s\n", name, err)
		return
	}
	if r.IsNew {
		fmt.Println("创建目录成功:", name)
	}

}
