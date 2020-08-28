package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan/apiutil"
	"path"
	"strings"
)

func RunRename(oldName string, newName string) {
	if oldName == "" {
		fmt.Println("请指定命名文件")
		return
	}
	if newName == "" {
		fmt.Println("请指定文件新名称")
		return
	}
	activeUser := GetActiveUser()
	oldName = activeUser.PathJoin(strings.TrimSpace(oldName))
	newName = activeUser.PathJoin(strings.TrimSpace(newName))
	if path.Dir(oldName) != path.Dir(newName) {
		fmt.Println("只能命名同一个目录的文件")
		return
	}
	if !apiutil.CheckFileNameValid(path.Base(newName)) {
		fmt.Println("文件名不能包含特殊字符：" + apiutil.FileNameSpecialChars)
		return
	}

	fileId := ""
	r, err := GetActivePanClient().FileInfoByPath(activeUser.PathJoin(oldName))
	if err != nil {
		fmt.Printf("原文件不存在： %s, %s\n", oldName, err)
		return
	}
	fileId = r.FileId

	b, e := activeUser.PanClient().Rename(fileId, path.Base(newName))
	if e != nil {
		fmt.Println(e.Err)
		return
	}
	if !b {
		fmt.Println("重命名文件失败")
		return
	}
	fmt.Printf("重命名文件成功：%s -> %s\n", path.Base(oldName), path.Base(newName))
}
