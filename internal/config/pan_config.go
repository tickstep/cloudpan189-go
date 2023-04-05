// Copyright (c) 2020 tickstep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package config

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/library/homedir"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil/jsonhelper"
	"github.com/tickstep/library-go/logger"
	"github.com/tickstep/library-go/requester"
)

const (
	// EnvVerbose 启用调试环境变量
	EnvVerbose = "CLOUD189_VERBOSE"
	// EnvConfigDir 配置路径环境变量
	EnvConfigDir = "CLOUD189_CONFIG_DIR"
	// ConfigName 配置文件名
	ConfigName = "cloud189_config.json"
	// ConfigVersion 配置文件版本
	ConfigVersion string = "1.0"

	// DefaultFileUploadParallelNum 默认的文件上传并发数量
	DefaultFileUploadParallelNum = 10

	// MaxFileUploadParallelNum 最大文件上传并发数量
	MaxFileUploadParallelNum = 20

	// DefaultFileDownloadParallelNum 默认的文件下载并发数量
	DefaultFileDownloadParallelNum = 5

	// MaxFileDownloadParallelNum 最大文件下载并发数量
	MaxFileDownloadParallelNum = 20
)

var (
	CmdConfigVerbose = logger.New("CONFIG", EnvVerbose)
	configFilePath   = filepath.Join(GetConfigDir(), ConfigName)

	// Config 配置信息, 由外部调用
	Config = NewConfig(configFilePath)

	AppVersion string
)

type UpdateCheckInfo struct {
	PreferUpdateSrv string `json:"preferUpdateSrv"` // 优先更新服务器，github | tickstep
	LatestVer       string `json:"latestVer"`       // 最后检测到的版本
	CheckTime       int64  `json:"checkTime"`       // 最后检测的时间戳，单位为秒
}

// PanConfig 配置详情
type PanConfig struct {
	ConfigVer string `json:"configVer"`
	ActiveUID uint64 `json:"activeUID"`

	UserList PanUserList `json:"userList"`

	CacheSize           int `json:"cacheSize"`           // 下载缓存
	MaxDownloadParallel int `json:"maxDownloadParallel"` // 最大下载并发量，即同时下载文件最大数量
	MaxUploadParallel   int `json:"maxUploadParallel"`   // 最大上传并发量，即同时上传文件最大数量

	MaxDownloadRate int64 `json:"maxDownloadRate"` // 限制最大下载速度，单位 B/s, 即字节/每秒
	MaxUploadRate   int64 `json:"maxUploadRate"`   // 限制最大上传速度，单位 B/s, 即字节/每秒

	SaveDir string `json:"saveDir"` // 下载储存路径

	Proxy           string          `json:"proxy"`      // 代理
	LocalAddrs      string          `json:"localAddrs"` // 本地网卡地址
	UpdateCheckInfo UpdateCheckInfo `json:"updateCheckInfo"`

	configFilePath string
	configFile     *os.File
	fileMu         sync.Mutex
	activeUser     *PanUser
}

// NewConfig 返回 PanConfig 指针对象
func NewConfig(configFilePath string) *PanConfig {
	c := &PanConfig{
		configFilePath: configFilePath,
	}
	return c
}

// Init 初始化配置
func (c *PanConfig) Init() error {
	return c.init()
}

// Reload 从文件重载配置
func (c *PanConfig) Reload() error {
	return c.init()
}

// Close 关闭配置文件
func (c *PanConfig) Close() error {
	if c.configFile != nil {
		err := c.configFile.Close()
		c.configFile = nil
		return err
	}
	return nil
}

// Save 保存配置信息到配置文件
func (c *PanConfig) Save() error {
	// 检测配置项是否合法, 不合法则自动修复
	c.fix()

	err := c.lazyOpenConfigFile()
	if err != nil {
		return err
	}

	c.fileMu.Lock()
	defer c.fileMu.Unlock()

	data, err := jsoniter.MarshalIndent(c, "", " ")
	if err != nil {
		// json数据生成失败
		panic(err)
	}

	// 减掉多余的部分
	err = c.configFile.Truncate(int64(len(data)))
	if err != nil {
		return err
	}

	_, err = c.configFile.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	_, err = c.configFile.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (c *PanConfig) init() error {
	if c.configFilePath == "" {
		return ErrConfigFileNotExist
	}

	c.initDefaultConfig()
	err := c.loadConfigFromFile()
	if err != nil {
		return err
	}

	// 设置全局代理
	if c.Proxy != "" {
		requester.SetGlobalProxy(c.Proxy)
	}
	// 设置本地网卡地址
	if c.LocalAddrs != "" {
		requester.SetLocalTCPAddrList(strings.Split(c.LocalAddrs, ",")...)
	}

	return nil
}

// lazyOpenConfigFile 打开配置文件
func (c *PanConfig) lazyOpenConfigFile() (err error) {
	if c.configFile != nil {
		return nil
	}

	c.fileMu.Lock()
	os.MkdirAll(filepath.Dir(c.configFilePath), 0700)
	c.configFile, err = os.OpenFile(c.configFilePath, os.O_CREATE|os.O_RDWR, 0600)
	c.fileMu.Unlock()

	if err != nil {
		if os.IsPermission(err) {
			return ErrConfigFileNoPermission
		}
		if os.IsExist(err) {
			return ErrConfigFileNotExist
		}
		return err
	}
	return nil
}

// loadConfigFromFile 载入配置
func (c *PanConfig) loadConfigFromFile() (err error) {
	err = c.lazyOpenConfigFile()
	if err != nil {
		return err
	}

	// 未初始化
	info, err := c.configFile.Stat()
	if err != nil {
		return err
	}

	if info.Size() == 0 {
		err = c.Save()
		return err
	}

	c.fileMu.Lock()
	defer c.fileMu.Unlock()

	_, err = c.configFile.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	err = jsonhelper.UnmarshalData(c.configFile, c)
	if err != nil {
		return ErrConfigContentsParseError
	}
	return nil
}

func (c *PanConfig) initDefaultConfig() {
	// 设置默认的下载路径
	switch runtime.GOOS {
	case "windows":
		c.SaveDir = cmdutil.ExecutablePathJoin("Downloads")
	case "android":
		// TODO: 获取完整的的下载路径
		c.SaveDir = "/sdcard/Download"
	default:
		dataPath, ok := os.LookupEnv("HOME")
		if !ok {
			CmdConfigVerbose.Warn("Environment HOME not set")
			c.SaveDir = cmdutil.ExecutablePathJoin("Downloads")
		} else {
			c.SaveDir = filepath.Join(dataPath, "Downloads")
		}
	}
	c.ConfigVer = ConfigVersion
}

// GetConfigDir 获取配置路径
func GetConfigDir() string {
	// 按照以下顺序依次获取配置目录
	// 1.环境变量CLOUD189_CONFIG_DIR => 2. /etc/cloud189/ => 3. ~/.cloud189/ => 4.当前程序目录

	// 1. 从环境变量读取
	configDir, ok := os.LookupEnv(EnvConfigDir)
	if ok {
		if filepath.IsAbs(configDir) {
			logger.Verboseln("use config dir from CLOUD189_CONFIG_DIR env: ", configDir)
			return configDir
		}
		// 如果不是绝对路径, 从程序目录寻找
		configDir = cmdutil.ExecutablePathJoin(configDir)
		logger.Verboseln("use config dir from CLOUD189_CONFIG_DIR env: ", configDir)
		return configDir
	} else {
		// 2. /etc/cloud189/
		if runtime.GOOS == "linux" {
			cd := "/etc/cloud189"
			if IsFolderExist(cd) {
				logger.Verboseln("use config dir: ", cd)
				return cd
			}
		}

		// 3. ~/.cloud189/
		if runtime.GOOS == "linux" || runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
			cd, er := homedir.Expand("~/.cloud189")
			if er == nil {
				if IsFolderExist(cd) {
					logger.Verboseln("use config dir: ", cd)
					return cd
				}
			}
		}
	}

	// 4.当前程序所在目录
	configDir = cmdutil.ExecutablePathJoin("")
	logger.Verboseln("use config dir: ", configDir)
	return configDir
}

func (c *PanConfig) ActiveUser() *PanUser {
	if c.activeUser == nil {
		if c.UserList == nil {
			return nil
		}
		if c.ActiveUID == 0 {
			return nil
		}
		for _, u := range c.UserList {
			if u.UID == c.ActiveUID {
				if u.PanClient() == nil {
					// restore client
					user, err := SetupUserByCookie(&u.WebToken, &u.AppToken)
					if err != nil {
						logger.Verboseln("setup user error")
						return nil
					}
					u.panClient = user.panClient
					u.Nickname = user.Nickname

					// check workdir valid or not
					if u.ActiveFamilyId > 0 {
						fe, err1 := u.PanClient().AppFileInfoByPath(u.ActiveFamilyId, u.FamilyWorkdir)
						if err1 != nil {
							// default to root
							u.FamilyWorkdir = "/"
							u.FamilyWorkdirFileEntity = *cloudpan.NewAppFileEntityForRootDir()
						} else {
							u.FamilyWorkdirFileEntity = *fe
						}
					} else {
						fe, err1 := u.PanClient().AppFileInfoByPath(u.ActiveFamilyId, u.Workdir)
						if err1 != nil {
							// default to root
							u.Workdir = "/"
							u.WorkdirFileEntity = *cloudpan.NewAppFileEntityForRootDir()
						} else {
							u.WorkdirFileEntity = *fe
						}
					}
				}
				c.activeUser = u
				return u
			}
		}
		return &PanUser{}
	}
	return c.activeUser
}

func (c *PanConfig) SetActiveUser(user *PanUser) *PanUser {
	needToInsert := true
	for _, u := range c.UserList {
		if u.UID == user.UID {
			// update user info
			u.Nickname = user.Nickname
			u.Sex = user.Sex
			u.WebToken = user.WebToken
			u.AppToken = user.AppToken
			u.LoginUserName = user.LoginUserName
			u.LoginUserPassword = user.LoginUserPassword
			needToInsert = false
			break
		}
	}
	if needToInsert {
		// insert
		c.UserList = append(c.UserList, user)
	}

	// setup active user
	c.ActiveUID = user.UID
	// clear active user cache
	c.activeUser = nil
	// reload
	return c.ActiveUser()
}

func (c *PanConfig) fix() {

}

// NumLogins 获取登录的用户数量
func (c *PanConfig) NumLogins() int {
	return len(c.UserList)
}

// SwitchUser 切换登录用户
func (c *PanConfig) SwitchUser(uid uint64, username string) (*PanUser, error) {
	for _, u := range c.UserList {
		if u.UID == uid || u.AccountName == username {
			return c.SetActiveUser(u), nil
		}
	}
	return nil, fmt.Errorf("未找到指定的账号")
}

// DeleteUser 删除用户，并自动切换登录用户为用户列表第一个
func (c *PanConfig) DeleteUser(uid uint64) (*PanUser, error) {
	for idx, u := range c.UserList {
		if u.UID == uid {
			// delete user from user list
			c.UserList = append(c.UserList[:idx], c.UserList[idx+1:]...)
			c.ActiveUID = 0
			c.activeUser = nil
			if len(c.UserList) > 0 {
				c.SwitchUser(c.UserList[0].UID, "")
			}
			return u, nil
		}
	}
	return nil, fmt.Errorf("未找到指定的账号")
}

// HTTPClient 返回设置好的 HTTPClient
func (c *PanConfig) HTTPClient(ua string) *requester.HTTPClient {
	client := requester.NewHTTPClient()
	if ua != "" {
		client.SetUserAgent(ua)
	}
	return client
}
