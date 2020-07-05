package config

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"path"
)

type PanUser struct {
	UID      uint64
	Nickname string
	AccountName string
	Sex      string
	Workdir  string
	WorkdirFileEntity cloudpan.FileEntity

	CookieLoginUser string
	AppToken cloudpan.AppLoginToken
	panClient *cloudpan.PanClient
}

type PanUserList []*PanUser

func SetupUserByCookie(cookieLoginUser string) (user *PanUser, err error) {
	panClient := cloudpan.NewPanClient(cookieLoginUser)
	u := &PanUser{
		CookieLoginUser: cookieLoginUser,
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