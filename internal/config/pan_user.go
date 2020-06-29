package config

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan"
)

type PanUser struct {
	UID      uint64
	Nickname string
	AccountName string
	Sex      string
	Workdir  string

	CookieLoginUser string
	panClient *cloudpan.PanClient
}

type PanUserList []*PanUser

func SetupUserByCookie(cookieLoginUser string) (user *PanUser, err error) {
	panClient := cloudpan.NewPanClient(cookieLoginUser)
	u := &PanUser{
		CookieLoginUser: cookieLoginUser,
		panClient: panClient,
		Workdir: "/",
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

func (p *PanUser) PanClient() *cloudpan.PanClient {
	return p.panClient
}