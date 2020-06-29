package command

import (
	"github.com/tickstep/cloudpan189-go/cloudpan"
)

func RunGetUserInfo() (userInfo *cloudpan.UserInfo, error error) {
	return GetActivePanClient().GetUserInfo()
}
