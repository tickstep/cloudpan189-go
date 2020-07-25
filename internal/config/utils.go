package config

import (
	"github.com/olekukonko/tablewriter"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"github.com/tickstep/cloudpan189-go/library/converter"
	"strconv"
	"strings"
)

func (pl *PanUserList) String() string {
	builder := &strings.Builder{}

	tb := cmdtable.NewTable(builder)
	tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})
	tb.SetHeader([]string{"#", "uid", "用户名", "昵称", "性别"})

	for k, userInfo := range *pl {
		sex := "未知"
		if userInfo.Sex == "F" {
			sex = "女"
		} else if userInfo.Sex == "M" {
			sex = "男"
		}
		tb.Append([]string{strconv.Itoa(k), strconv.FormatUint(userInfo.UID, 10), userInfo.AccountName, userInfo.Nickname, sex})
	}

	tb.Render()

	return builder.String()
}

// AverageParallel 返回平均的下载最大并发量
func AverageParallel(parallel, downloadLoad int) int {
	if downloadLoad < 1 {
		return 1
	}

	p := parallel / downloadLoad
	if p < 1 {
		return 1
	}
	return p
}

func stripPerSecond(sizeStr string) string {
	i := strings.LastIndex(sizeStr, "/")
	if i < 0 {
		return sizeStr
	}
	return sizeStr[:i]
}

func showMaxRate(size int64) string {
	if size <= 0 {
		return "不限制"
	}
	return converter.ConvertFileSize(size, 2) + "/s"
}
