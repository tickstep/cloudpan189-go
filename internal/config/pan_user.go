package config

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"path"
	"path/filepath"
)

type PanUser struct {
	UID      uint64 `json:"uid"`
	Nickname string `json:"nickname"`
	AccountName string `json:"accountName"`
	Sex      string `json:"sex"`
	Workdir  string `json:"workdir"`
	WorkdirFileEntity cloudpan.FileEntity `json:"workdirFileEntity"`

	LoginUserName string `json:"loginUserName"`
	LoginUserPassword string `json:"loginUserPassword"`

	WebToken cloudpan.WebLoginToken `json:"webToken"`
	AppToken cloudpan.AppLoginToken `json:"appToken"`
	panClient *cloudpan.PanClient
}

type PanUserList []*PanUser

func SetupUserByCookie(webToken *cloudpan.WebLoginToken, appToken *cloudpan.AppLoginToken) (user *PanUser, err *apierror.ApiError) {
	tryRefreshWebToken := true

doLoginAct:
	panClient := cloudpan.NewPanClient(*webToken, *appToken)
	u := &PanUser{
		WebToken: *webToken,
		AppToken: *appToken,
		panClient: panClient,
		Workdir: "/",
		WorkdirFileEntity: *cloudpan.NewFileEntityForRootDir(),
	}

	// web api token maybe expired
	userInfo, err := panClient.GetUserInfo()
	if err != nil {
		if err.Code == apierror.ApiCodeTokenExpiredCode && appToken.SessionKey != "" && tryRefreshWebToken {
			tryRefreshWebToken = false
			webCookie := cloudpan.RefreshCookieToken(appToken.SessionKey)
			if webCookie != "" {
				webToken.CookieLoginUser = webCookie
				goto doLoginAct
			}
		}
		return nil, err
	}
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
		return nil, apierror.NewFailedApiError("cannot get user info, the token has expired")
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
		return nil, apierror.NewFailedApiError("cannot get user info, the token has expired")
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

// GetSavePath 根据提供的网盘文件路径 panpath, 返回本地储存路径,
// 返回绝对路径, 获取绝对路径出错时才返回相对路径...
func (pu *PanUser) GetSavePath(filePanPath string) string {
	dirStr := filepath.Join(Config.SaveDir, fmt.Sprintf("%d", pu.UID), filePanPath)
	dir, err := filepath.Abs(dirStr)
	if err != nil {
		dir = filepath.Clean(dirStr)
	}
	return dir
}