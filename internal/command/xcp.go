package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
)

type (
	FileSourceType string
)

const (
	// 个人云文件
	PersonCloud FileSourceType = "person"

	// 家庭云文件
	FamilyCloud FileSourceType = "family"
)

// RunXCopy 执行移动文件/目录
func RunXCopy(source FileSourceType, familyId int64, paths ...string) {
	activeUser := GetActiveUser()

	// use the first family as default
	if familyId == 0 {
		familyResult,err := activeUser.PanClient().AppFamilyGetFamilyList()
		if err != nil {
			fmt.Println("获取家庭列表失败")
			return
		}
		familyId = familyResult.FamilyInfoList[0].FamilyId
	}

	var opFileList []*cloudpan.AppFileEntity
	var failedPaths []string
	var err error
	switch source {
	case FamilyCloud:
		opFileList, failedPaths, err = GetAppFileInfoByPaths(familyId, paths...)
		break
	case PersonCloud:
		opFileList, failedPaths, err = GetAppFileInfoByPaths(0, paths...)
		break
	default:
		fmt.Println("不支持的云类型")
		return
	}

	if err !=  nil {
		fmt.Println(err)
		return
	}
	if opFileList == nil || len(opFileList) == 0 {
		fmt.Println("没有有效的文件可复制")
		return
	}

	fileIdList := []string{}
	for _,fi := range opFileList {
		fileIdList = append(fileIdList, fi.FileId)
	}

	switch source {
	case FamilyCloud:
		// copy to person cloud
		_,e1 := activeUser.PanClient().AppFamilySaveFileToPersonCloud(familyId, fileIdList)
		if e1 != nil {
			if e1.ErrCode() == apierror.ApiCodeFileAlreadyExisted {
				fmt.Println("复制失败，个人云已经存在对应的文件")
			} else {
				fmt.Println("复制文件到个人云失败")
			}
			return
		}
		break
	case PersonCloud:
		// copy to family cloud
		_,e1 := activeUser.PanClient().AppSaveFileToFamilyCloud(familyId, fileIdList)
		if e1 != nil {
			if e1.ErrCode() == apierror.ApiCodeFileAlreadyExisted {
				fmt.Println("复制失败，家庭云已经存在对应的文件")
			} else {
				fmt.Println("复制文件到家庭云失败")
			}
			return
		}
		break
	default:
		fmt.Println("不支持的云类型")
		return
	}

	if len(failedPaths) > 0 {
		fmt.Println("以下文件复制失败：")
		for _,f := range failedPaths {
			fmt.Println(f)
		}
		fmt.Println("")
	}

	switch source {
	case FamilyCloud:
		// copy to person cloud
		fmt.Println("成功复制以下文件到个人云目录 /来自家庭共享")
		for _,fi := range opFileList {
			fmt.Println(fi.Path)
		}
		break
	case PersonCloud:
		// copy to family cloud
		fmt.Println("成功复制以下文件到家庭云根目录")
		for _,fi := range opFileList {
			fmt.Println(fi.Path)
		}
		break
	default:
		fmt.Println("不支持的云类型")
		return
	}
}