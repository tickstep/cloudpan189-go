package command

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"strconv"
	"strings"
)

func RunSwitchFamilyList(targetFamilyId int64)  {
	var activeFamilyInfo *cloudpan.AppFamilyInfo = nil
	familyList,renderStr := getFamilyOptionList()

	if familyList == nil || len(familyList) == 0 {
		fmt.Println("切换家庭云失败")
		return
	}

	if targetFamilyId < 0 {
		// show option list
		fmt.Println(renderStr)

		// 提示输入 index
		var index string
		fmt.Printf("输入要切换的家庭云 # 值 > ")
		_, err := fmt.Scanln(&index)
		if err != nil {
			return
		}

		if n, err := strconv.Atoi(index); err == nil && n >= 0 && n < len(familyList) {
			activeFamilyInfo = familyList[n]
		} else {
			fmt.Printf("切换家庭云失败, 请检查 # 值是否正确\n")
			return
		}
	} else {
		// 直接切换
		for _,familyInfo := range familyList {
			if familyInfo.FamilyId == targetFamilyId {
				activeFamilyInfo = familyInfo
				break
			}
		}
	}

	if activeFamilyInfo == nil {
		fmt.Printf("切换家庭云失败\n")
		return
	}

	config.Config.ActiveUser().ActiveFamilyId = activeFamilyInfo.FamilyId
	config.Config.ActiveUser().ActiveFamilyInfo = *activeFamilyInfo
	fmt.Printf("切换家庭云：%s\n", activeFamilyInfo.RemarkName)
}

func getFamilyOptionList() ([]*cloudpan.AppFamilyInfo, string) {
	activeUser := config.Config.ActiveUser()

	familyResult,err := activeUser.PanClient().AppFamilyGetFamilyList()
	if err != nil {
		fmt.Println("获取家庭列表失败")
		return nil, ""
	}
	t := []*cloudpan.AppFamilyInfo{}
	personCloud := &cloudpan.AppFamilyInfo{
		FamilyId: 0,
		RemarkName: "个人云",
		CreateTime: "-",
	}
	t = append(t, personCloud)
	t = append(t, familyResult.FamilyInfoList...)
	familyList := t
	builder := &strings.Builder{}
	tb := cmdtable.NewTable(builder)
	tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})
	tb.SetHeader([]string{"#", "family_id", "家庭云名", "创建日期"})

	for k, familyInfo := range familyList {
		tb.Append([]string{strconv.Itoa(k), strconv.FormatInt(familyInfo.FamilyId, 10), familyInfo.RemarkName, familyInfo.CreateTime})
	}
	tb.Render()
	return familyList, builder.String()
}
