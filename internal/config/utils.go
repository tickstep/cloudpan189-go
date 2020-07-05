package config

import (
	"github.com/olekukonko/tablewriter"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
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

