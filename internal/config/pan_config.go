package config

import (
	"github.com/json-iterator/go"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil/jsonhelper"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	// EnvConfigDir 配置路径环境变量
	EnvConfigDir = "CLOUD189_CONFIG_DIR"
	// ConfigName 配置文件名
	ConfigName = "cloud189_config.json"
)

var (
	CmdConfigVerbose = logger.New("CONFIG")
	configFilePath   = filepath.Join(GetConfigDir(), ConfigName)

	// Config 配置信息, 由外部调用
	Config = NewConfig(configFilePath)
)

// PanConfig 配置详情
type PanConfig struct {
	ActiveUID uint64
	UserList  PanUserList

	SaveDir        string // 下载储存路径

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
}

// GetConfigDir 获取配置路径
func GetConfigDir() string {
	// 从环境变量读取
	configDir, ok := os.LookupEnv(EnvConfigDir)
	if ok {
		if filepath.IsAbs(configDir) {
			return configDir
		}
		// 如果不是绝对路径, 从程序目录寻找
		return cmdutil.ExecutablePathJoin(configDir)
	}
	return cmdutil.ExecutablePathJoin(configDir)
}

func (c *PanConfig) ActiveUser() *PanUser {
	if c.activeUser == nil {
		if c.UserList == nil {
			return nil
		}
		for _, u := range c.UserList {
			if u.UID == c.ActiveUID {
				c.activeUser = u
				if c.activeUser.PanClient() == nil {
					// restore client
					user, err := SetupUserByCookie(c.activeUser.CookieLoginUser)
					if err != nil {
						return nil
					}
					return c.SetActiveUser(user)
				}
				return u
			}
		}
		return &PanUser{}
	}
	return c.activeUser
}

func (c *PanConfig) SetActiveUser(user *PanUser) *PanUser {
	// remove old user
	for i, u := range c.UserList {
		if u.UID == user.UID {
			c.UserList = append(c.UserList[:i], c.UserList[i+1:]...)
			break
		}
	}
	// insert new old
	c.UserList = append(c.UserList, user)
	c.activeUser = user
	c.ActiveUID = user.UID
	return c.activeUser
}

func (c *PanConfig) fix() {

}