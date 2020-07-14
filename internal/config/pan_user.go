package config

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"path"
)

type PanUser struct {
	UID      uint64 `json:"uid"`
	Nickname string `json:"nickname"`
	AccountName string `json:"account_name"`
	Sex      string `json:"sex"`
	Workdir  string `json:"workdir"`
	WorkdirFileEntity cloudpan.FileEntity `json:"workdir_file_entity"`

	WebToken cloudpan.WebLoginToken `json:"web_token"`
	AppToken cloudpan.AppLoginToken `json:"app_token"`
	panClient *cloudpan.PanClient
}

type PanUserList []*PanUser

func SetupUserByCookie(webToken cloudpan.WebLoginToken, appToken cloudpan.AppLoginToken) (user *PanUser, err error) {
	panClient := cloudpan.NewPanClient(webToken, appToken)
	u := &PanUser{
		WebToken: webToken,
		AppToken: appToken,
		panClient: panClient,
		Workdir: "/",
		WorkdirFileEntity: *cloudpan.NewFileEntityForRootDir(),
	}

	userInfo, err := panClient.GetUserInfo()
	name := "Unknown"
	if userInfo != nil {
		name = userInfo.Nickname
		if name == "" {
			name = userInfo.UserAccount
		}

		// update cloudUser
		u.UID = userInfo.UserId
		u.AccountName = userInfo.UserAccount
	} else {
		// error, maybe the token has expired
		return nil, fmt.Errorf("cannot get user info, the token has expired")
	}
	u.Nickname = name

	userDetailInfo, err := panClient.GetUserDetailInfo()
	if userDetailInfo != nil {
		if userDetailInfo.Gender == "F" {
			u.Sex = "F"
		} else if userDetailInfo.Gender == "M" {
			u.Sex = "M"
		} else {
			u.Sex = "U"
		}
	} else {
		// error, maybe the token has expired
		return nil, fmt.Errorf("cannot get user info, the token has expired")
	}

	return u, nil
}

func (pu *PanUser) PanClient() *cloudpan.PanClient {
	return pu.panClient
}

// PathJoin 合并工作目录和相对路径p, 若p为绝对路径则忽略
func (pu *PanUser) PathJoin(p string) string {
	if path.IsAbs(p) {
		return p
	}
	return path.Join(pu.Workdir, p)
}

func (pu *PanUser) FreshWorkdirInfo() {
	fe, err := pu.PanClient().FileInfoById(pu.WorkdirFileEntity.FileId)
	if err != nil {
		logger.Verboseln("刷新工作目录信息失败")
		return
	}
	pu.WorkdirFileEntity = *fe
}