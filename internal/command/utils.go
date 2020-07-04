package command

import "strings"

const (
	FileNameSpecialChars = "\\/:*?\"<>|"
)

// CheckFileNameValid 检测文件名是否有效，包含特殊字符则无效
func CheckFileNameValid(name string) bool {
	if name == "" {
		return true
	}
	return !strings.ContainsAny(name, FileNameSpecialChars)
}
