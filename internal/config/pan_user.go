package config

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apiweb"
)

type PanUser struct {
	UID      uint64
	Nickname string
	AccountName string
	Sex      string
	Workdir  string

	CookieLoginUser string
	panClient *PanClient
}

type PanUserList []*PanUser

func SetupUserByCookie(CookieLoginUser string) (user *PanUser, err error) {
	panClient := NewPanClient(CookieLoginUser)
	u := &PanUser{
		CookieLoginUser: CookieLoginUser,
		panClient: panClient,
		Workdir: "/",
	}

	userInfo, err := apiweb.GetUserInfo(panClient)
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

	userDetailInfo, err := apiweb.GetUserDetailInfo(panClient)
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

func (p *PanUser) PanClient() *PanClient {
	return p.panClient
}