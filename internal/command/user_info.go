package command

import (
	"github.com/tickstep/cloudpan189-go/cloudpan/apiweb"
	"github.com/tickstep/cloudpan189-go/internal/config"
)

func RunGetUserInfo() (userInfo *apiweb.UserInfo, error error) {
	return apiweb.GetUserInfo(config.Config.ActiveUser().PanClient())
}
