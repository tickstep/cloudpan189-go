package command

import (
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/internal/config"
)

func GetActivePanClient() *cloudpan.PanClient {
	return config.Config.ActiveUser().PanClient()
}

func GetActiveUser() *config.PanUser {
	return config.Config.ActiveUser()
}