package config

import "github.com/tickstep/cloudpan189-go/cloudpan/apiweb"

type PanUser struct {
	UID  uint64
	Name string
	Workdir string

	CookieLoginUser string
	panClient *PanClient
}

type PanUserList []*PanUser

func SetupUserByCookie(CookieLoginUser string) (user *PanUser) {
	panClient := NewPanClient(CookieLoginUser)
	u := &PanUser{
		CookieLoginUser: CookieLoginUser,
		panClient: panClient,
		Workdir: "/",
	}

	userInfo, err := apiweb.GetUserInfo(panClient)
	name := "Unknown"
	if err == nil && userInfo != nil {
		name = userInfo.Nickname
		if name == "" {
			name = userInfo.UserAccount
		}

		// update cloudUser
		u.UID = userInfo.UserId
	}
	u.Name = name
	return u
}

func (p *PanUser) PanClient() *PanClient {
	return p.panClient
}