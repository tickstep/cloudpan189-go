package main

import (
	"fmt"
	"github.com/peterh/liner"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdliner"
	"github.com/tickstep/cloudpan189-go/cmder/cmdliner/args"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil/escaper"
	"github.com/tickstep/cloudpan189-go/internal/command"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/cloudpan189-go/internal/functions/pandownload"
	"github.com/tickstep/cloudpan189-go/internal/panupdate"
	"github.com/tickstep/library-go/converter"
	"github.com/tickstep/library-go/logger"
	"github.com/urfave/cli"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	// NameShortDisplayNum 文件名缩略显示长度
	NameShortDisplayNum = 16
)

var (
	// Version 版本号
	Version = "v0.0.8-dev"

	saveConfigMutex *sync.Mutex = new(sync.Mutex)

	historyFilePath = filepath.Join(config.GetConfigDir(), "cloud189_command_history.txt")
	reloadFn        = func(c *cli.Context) error {
		err := config.Config.Reload()
		if err != nil {
			fmt.Printf("重载配置错误: %s\n", err)
		}
		return nil
	}
	saveFunc = func(c *cli.Context) error {
		saveConfigMutex.Lock()
		defer saveConfigMutex.Unlock()
		err := config.Config.Save()
		if err != nil {
			fmt.Printf("保存配置错误: %s\n", err)
		}
		return nil
	}

	isCli bool
)

func init() {
	config.AppVersion = Version
	cmdutil.ChWorkDir()

	err := config.Config.Init()
	switch err {
	case nil:
	case config.ErrConfigFileNoPermission, config.ErrConfigContentsParseError:
		fmt.Fprintf(os.Stderr, "FATAL ERROR: config file error: %s\n", err)
		os.Exit(1)
	default:
		fmt.Printf("WARNING: config init error: %s\n", err)
	}
}

func tryLogin() *config.PanUser {
	// can do automatically login?
	for _, u := range config.Config.UserList {
		if u.UID == config.Config.ActiveUID {
			// login
			_, _, webToken, appToken, err := command.RunLogin(config.DecryptString(u.LoginUserName), config.DecryptString(u.LoginUserPassword))
			if err != nil {
				logger.Verboseln("automatically login error")
				break
			}
			// success
			u.WebToken = webToken
			u.AppToken = appToken

			// save
			saveFunc(nil)
			// reload
			reloadFn(nil)
			return config.Config.ActiveUser()
		}
	}
	return nil
}

func checkLoginExpiredAndRelogin() {
	reloadFn(nil)
	activeUser := config.Config.ActiveUser()
	if activeUser == nil {
		// maybe expired, try to login
		tryLogin()
	}
	saveFunc(nil)
}

func parseFamilyId(c *cli.Context) int64 {
	familyId := config.Config.ActiveUser().ActiveFamilyId
	if c.IsSet("familyId") {
		fid,errfi := strconv.ParseInt(c.String("familyId"), 10, 64)
		if errfi != nil {
			familyId = 0
		} else {
			familyId = fid
		}
	}
	return familyId
}

func main() {
	defer config.Config.Close()

	// check & relogin
	checkLoginExpiredAndRelogin()

	app := cli.NewApp()
	app.Name = "cloudpan189-go"
	app.Version = Version
	app.Author = "tickstep/cloudpan189-go: https://github.com/tickstep/cloudpan189-go"
	app.Copyright = "(c) 2020 tickstep."
	app.Usage = "天翼云盘客户端 for " + runtime.GOOS + "/" + runtime.GOARCH
	app.Description = `cloudpan189-go 使用Go语言编写的天翼云盘命令行客户端, 为操作天翼云盘, 提供实用功能.
	具体功能, 参见 COMMANDS 列表

	------------------------------------------------------------------------------
	前往 https://github.com/tickstep/cloudpan189-go 以获取更多帮助信息!
	前往 https://github.com/tickstep/cloudpan189-go/releases 以获取程序更新信息!
	------------------------------------------------------------------------------

	交流反馈:
		提交Issue: https://github.com/tickstep/cloudpan189-go/issues
		联系邮箱: tickstep@outlook.com`

	// 全局options
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "启用调试",
			EnvVar:      config.EnvVerbose,
			Destination: &logger.IsVerbose,
		},
	}

	// 进入交互CLI命令行界面
	app.Action = func(c *cli.Context) {
		if c.NArg() != 0 {
			fmt.Printf("未找到命令: %s\n运行命令 %s help 获取帮助\n", c.Args().Get(0), app.Name)
			return
		}

		isCli = true
		logger.Verbosef("VERBOSE: 这是一条调试信息\n\n")

		var (
			line = cmdliner.NewLiner()
			err  error
		)

		line.History, err = cmdliner.NewLineHistory(historyFilePath)
		if err != nil {
			fmt.Printf("警告: 读取历史命令文件错误, %s\n", err)
		}

		line.ReadHistory()
		defer func() {
			line.DoWriteHistory()
			line.Close()
		}()

		// tab 自动补全命令
		line.State.SetCompleter(func(line string) (s []string) {
			var (
				lineArgs                   = args.Parse(line)
				numArgs                    = len(lineArgs)
				acceptCompleteFileCommands = []string{
					"cd", "cp", "xcp", "download", "ls", "mkdir", "mv", "pwd", "rename", "rm", "share", "upload", "login", "loglist", "logout",
					"clear", "quit", "exit", "quota", "who", "sign", "update", "who", "su", "config",
					"family",
				}
				closed = strings.LastIndex(line, " ") == len(line)-1
			)

			for _, cmd := range app.Commands {
				for _, name := range cmd.Names() {
					if !strings.HasPrefix(name, line) {
						continue
					}

					s = append(s, name+" ")
				}
			}

			switch numArgs {
			case 0:
				return
			case 1:
				if !closed {
					return
				}
			}

			thisCmd := app.Command(lineArgs[0])
			if thisCmd == nil {
				return
			}

			if !cmdutil.ContainsString(acceptCompleteFileCommands, thisCmd.FullName()) {
				return
			}

			var (
				activeUser  = config.Config.ActiveUser()
				runeFunc    = unicode.IsSpace
				//cmdRuneFunc = func(r rune) bool {
				//	switch r {
				//	case '\'', '"':
				//		return true
				//	}
				//	return unicode.IsSpace(r)
				//}
				targetPath string
			)

			if !closed {
				targetPath = lineArgs[numArgs-1]
				escaper.EscapeStringsByRuneFunc(lineArgs[:numArgs-1], runeFunc) // 转义
			} else {
				escaper.EscapeStringsByRuneFunc(lineArgs, runeFunc)
			}

			switch {
			case targetPath == "." || strings.HasSuffix(targetPath, "/."):
				s = append(s, line+"/")
				return
			case targetPath == ".." || strings.HasSuffix(targetPath, "/.."):
				s = append(s, line+"/")
				return
			}

			var (
				targetDir string
				isAbs     = path.IsAbs(targetPath)
				isDir     = strings.LastIndex(targetPath, "/") == len(targetPath)-1
			)

			if isAbs {
				targetDir = path.Dir(targetPath)
			} else {
				targetDir = path.Join(activeUser.Workdir, targetPath)
				if !isDir {
					targetDir = path.Dir(targetDir)
				}
			}

			return
		})

		fmt.Printf("提示: 方向键上下可切换历史命令.\n")
		fmt.Printf("提示: Ctrl + A / E 跳转命令 首 / 尾.\n")
		fmt.Printf("提示: 输入 help 获取帮助.\n")

		// check update
		reloadFn(c)
		if config.Config.UpdateCheckInfo.LatestVer != "" {
			if config.Config.UpdateCheckInfo.LatestVer > config.AppVersion {
				fmt.Printf("\n当前的软件版本为：%s， 现在有新版本 %s 可供更新，强烈推荐进行更新！（可以输入 update 命令进行更新）\n\n",
					config.AppVersion, config.Config.UpdateCheckInfo.LatestVer)
			}
		}
		go func() {
			latestCheckTime := config.Config.UpdateCheckInfo.CheckTime
			nowTime := time.Now().Unix()
			secsOf12Hour := int64(43200)
			if (nowTime - latestCheckTime) > secsOf12Hour {
				releaseInfo := panupdate.GetLatestReleaseInfo(false)
				if releaseInfo == nil {
					logger.Verboseln("获取版本信息失败!\n")
					return
				}
				config.Config.UpdateCheckInfo.LatestVer = releaseInfo.TagName
				config.Config.UpdateCheckInfo.CheckTime = nowTime

				// save
				saveFunc(c)
			}
		}()

		for {
			var (
				prompt     string
				activeUser = config.Config.ActiveUser()
			)

			if activeUser == nil {
				activeUser = tryLogin()
			}

			if activeUser != nil && activeUser.Nickname != "" {
				// 格式: cloudpan189-go:<工作目录> <UserName>$
				// 工作目录太长时, 会自动缩略
				if command.IsFamilyCloud(activeUser.ActiveFamilyId) {
					prompt = app.Name + ":" + converter.ShortDisplay(path.Base(activeUser.FamilyWorkdir), NameShortDisplayNum) + " " + activeUser.Nickname + "(" + command.GetFamilyCloudMark(activeUser.ActiveFamilyId) + ")$ "
				} else {
					prompt = app.Name + ":" + converter.ShortDisplay(path.Base(activeUser.Workdir), NameShortDisplayNum) + " " + activeUser.Nickname + "$ "
				}
			} else {
				// cloudpan189-go >
				prompt = app.Name + " > "
			}

			commandLine, err := line.State.Prompt(prompt)
			switch err {
			case liner.ErrPromptAborted:
				return
			case nil:
				// continue
			default:
				fmt.Println(err)
				return
			}

			line.State.AppendHistory(commandLine)

			cmdArgs := args.Parse(commandLine)
			if len(cmdArgs) == 0 {
				continue
			}

			s := []string{os.Args[0]}
			s = append(s, cmdArgs...)

			// 恢复原始终端状态
			// 防止运行命令时程序被结束, 终端出现异常
			line.Pause()
			c.App.Run(s)
			line.Resume()
		}
	}

	// 命令配置和对应的处理func
	app.Commands = []cli.Command{
		// 登录账号 login
		{
			Name:  "login",
			Usage: "登录天翼云盘账号",
			Description: `
	示例:
		cloudpan189-go login
		cloudpan189-go login -username=tickstep -password=123xxx

	常规登录:
		按提示一步一步来即可.
`,
			Category: "天翼云盘账号",
			Before:   reloadFn, // 每次进行登录动作的时候需要调用刷新配置
			After:    saveFunc, // 登录完成需要调用保存配置
			Action: func(c *cli.Context) error {
				appToken := cloudpan.AppLoginToken{}
				webToken := cloudpan.WebLoginToken{}
				username := ""
				passowrd := ""
				if c.IsSet("COOKIE_LOGIN_USER") {
					webToken.CookieLoginUser = c.String("COOKIE_LOGIN_USER")
				} else if c.NArg() == 0 {
					var err error
					username, passowrd, webToken, appToken, err = command.RunLogin(c.String("username"), c.String("password"))
					if err != nil {
						fmt.Println(err)
						return err
					}
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				cloudUser, _ := config.SetupUserByCookie(&webToken, &appToken)
				// save username / password
				cloudUser.LoginUserName = config.EncryptString(username)
				cloudUser.LoginUserPassword = config.EncryptString(passowrd)
				config.Config.SetActiveUser(cloudUser)
				fmt.Println("天翼帐号登录成功: ", cloudUser.Nickname)
				return nil
			},
			// 命令的附加options参数说明，使用 help login 命令即可查看
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "username",
					Usage: "登录天翼帐号的用户名(手机号/邮箱/别名)",
				},
				cli.StringFlag{
					Name:  "password",
					Usage: "登录天翼帐号的用户密码",
				},
				// 暂不支持
				// cloudpan189-go login -COOKIE_LOGIN_USER=8B12CBBCE89CA8DFC3445985B63B511B5E7EC7...
				//cli.StringFlag{
				//	Name:  "COOKIE_LOGIN_USER",
				//	Usage: "使用 COOKIE_LOGIN_USER cookie来登录帐号",
				//},
			},
		},
		// 退出登录帐号 logout
		{
			Name:        "logout",
			Usage:       "退出天翼帐号",
			Description: "退出当前登录的帐号",
			Category:    "天翼云盘账号",
			Before:      reloadFn,
			After:       saveFunc,
			Action: func(c *cli.Context) error {
				if config.Config.NumLogins() == 0 {
					fmt.Println("未设置任何帐号, 不能退出")
					return nil
				}

				var (
					confirm    string
					activeUser = config.Config.ActiveUser()
				)

				if !c.Bool("y") {
					fmt.Printf("确认退出当前帐号: %s ? (y/n) > ", activeUser.Nickname)
					_, err := fmt.Scanln(&confirm)
					if err != nil || (confirm != "y" && confirm != "Y") {
						return err
					}
				}

				deletedUser, err := config.Config.DeleteUser(activeUser.UID)
				if err != nil {
					fmt.Printf("退出用户 %s, 失败, 错误: %s\n", activeUser.Nickname, err)
				}

				fmt.Printf("退出用户成功: %s\n", deletedUser.Nickname)
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "y",
					Usage: "确认退出帐号",
				},
			},
		},
		// 列出帐号列表 loglist
		{
			Name:        "loglist",
			Usage:       "列出帐号列表",
			Description: "列出所有已登录的天翼帐号",
			Category:    "天翼云盘账号",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				fmt.Println(config.Config.UserList.String())
				return nil
			},
		},
		// 切换天翼帐号 su
		{
			Name:  "su",
			Usage: "切换天翼帐号",
			Description: `
	切换已登录的天翼帐号:
	如果运行该条命令没有提供参数, 程序将会列出所有的帐号, 供选择切换.

	示例:
	cloudpan189-go su
	cloudpan189-go su <uid or name>
`,
			Category: "天翼云盘账号",
			Before:   reloadFn,
			After:    saveFunc,
			Action: func(c *cli.Context) error {
				if c.NArg() >= 2 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				numLogins := config.Config.NumLogins()

				if numLogins == 0 {
					fmt.Printf("未设置任何帐号, 不能切换\n")
					return nil
				}

				var (
					inputData = c.Args().Get(0)
					uid       uint64
				)

				if c.NArg() == 1 {
					// 直接切换
					uid, _ = strconv.ParseUint(inputData, 10, 64)
				} else if c.NArg() == 0 {
					// 输出所有帐号供选择切换
					cli.HandleAction(app.Command("loglist").Action, c)

					// 提示输入 index
					var index string
					fmt.Printf("输入要切换帐号的 # 值 > ")
					_, err := fmt.Scanln(&index)
					if err != nil {
						return nil
					}

					if n, err := strconv.Atoi(index); err == nil && n >= 0 && n < numLogins {
						uid = config.Config.UserList[n].UID
					} else {
						fmt.Printf("切换用户失败, 请检查 # 值是否正确\n")
						return nil
					}
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
				}

				switchedUser, err := config.Config.SwitchUser(uid, inputData)
				if err != nil {
					fmt.Printf("切换用户失败, %s\n", err)
					return nil
				}

				if switchedUser == nil {
					switchedUser = tryLogin()
				}

				if switchedUser != nil {
					fmt.Printf("切换用户: %s\n", switchedUser.Nickname)
				} else {
					fmt.Printf("切换用户失败\n")
				}

				return nil
			},
		},
		// 获取当前帐号 who
		{
			Name:        "who",
			Usage:       "获取当前帐号",
			Description: "获取当前帐号的信息",
			Category:    "天翼云盘账号",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				if config.Config.ActiveUser() == nil {
					fmt.Println("未登录账号")
					return nil
				}
				activeUser := config.Config.ActiveUser()
				gender := "未知"
				if activeUser.Sex == "F" {
					gender = "女"
				} else if activeUser.Sex == "M" {
					gender = "男"
				}
				cloudName := "个人云"
				if config.Config.ActiveUser().ActiveFamilyId > 0 {
					cloudName = "家庭云(" + config.Config.ActiveUser().ActiveFamilyInfo.RemarkName + ")"
				}
				fmt.Printf("当前帐号 uid: %d, 昵称: %s, 用户名: %s, 性别: %s, 云：%s\n", activeUser.UID, activeUser.Nickname, activeUser.AccountName, gender, cloudName)
				return nil
			},
		},
		// 切换家庭云 family
		{
			Name:  "family",
			Usage: "切换天翼家庭云/个人云",
			Description: `
	切换已登录的天翼帐号的家庭云和个人云:
	如果运行该条命令没有提供参数, 程序将会列出所有的家庭云, 供选择切换.

	示例:
	cloudpan189-go family
	cloudpan189-go family <familyId>
`,
			Category: "天翼云盘账号",
			Before:   reloadFn,
			After:    saveFunc,
			Action: func(c *cli.Context) error {
				inputData := c.Args().Get(0)
				targetFamilyId := int64(-1)
				if inputData != "" && len(inputData) > 0 {
					targetFamilyId,_ = strconv.ParseInt(inputData, 10, 0)
				}
				command.RunSwitchFamilyList(targetFamilyId)
				return nil
			},
		},
		// 获取当前帐号空间配额 quota
		{
			Name:        "quota",
			Usage:       "获取当前帐号空间配额",
			Description: "获取网盘的总储存空间, 和已使用的储存空间",
			Category:    "天翼云盘账号",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				if config.Config.ActiveUser() == nil {
					fmt.Println("未登录账号")
					return nil
				}
				q, err := command.RunGetQuotaInfo()
				if err == nil {
					fmt.Printf("账号: %s, uid: %d, 个人空间总额: %s, 个人空间已使用: %s, 比率: %f%%\n",
						config.Config.ActiveUser().Nickname, config.Config.ActiveUser().UID,
						converter.ConvertFileSize(q.Quota, 2), converter.ConvertFileSize(q.UsedSize, 2),
						100*float64(q.UsedSize)/float64(q.Quota))
				}
				return nil
			},
		},
		// 用户签到 sign
		{
			Name:        "sign",
			Usage:       "用户签到",
			Description: "当前帐号进行签到",
			Category:    "天翼云盘账号",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				if config.Config.ActiveUser() == nil {
					fmt.Println("未登录账号")
					return nil
				}
				command.RunUserSign()
				return nil
			},
		},
		// 切换工作目录 cd
		{
			Name:     "cd",
			Category: "天翼云盘",
			Usage:    "切换工作目录",
			Description: `
	cloudpan189-go cd <目录, 绝对路径或相对路径>

	示例:

	切换 /我的资源 工作目录:
	cloudpan189-go cd /我的资源

	切换上级目录:
	cloudpan189-go cd ..

	切换根目录:
	cloudpan189-go cd /
`,
			Before: reloadFn,
			After:  saveFunc,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				if config.Config.ActiveUser() == nil {
					fmt.Println("未登录账号")
					return nil
				}
				command.RunChangeDirectory(parseFamilyId(c), c.Args().Get(0))
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "familyId",
					Usage: "家庭云ID",
					Value: "",
				},
			},
		},
		// 输出工作目录 pwd
		{
			Name:      "pwd",
			Usage:     "输出工作目录",
			UsageText: app.Name + " pwd",
			Category:  "天翼云盘",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				if config.Config.ActiveUser() == nil {
					fmt.Println("未登录账号")
					return nil
				}
				if command.IsFamilyCloud(config.Config.ActiveUser().ActiveFamilyId) {
					fmt.Println(config.Config.ActiveUser().FamilyWorkdir)
				} else {
					fmt.Println(config.Config.ActiveUser().Workdir)
				}
				return nil
			},
		},
		// 列出目录 ls
		{
			Name:      "ls",
			Aliases:   []string{"l", "ll"},
			Usage:     "列出目录",
			UsageText: app.Name + " ls <目录>",
			Description: `
	列出当前工作目录内的文件和目录, 或指定目录内的文件和目录

	示例:

	列出 我的资源 内的文件和目录
	cloudpan189-go ls 我的资源

	绝对路径
	cloudpan189-go ls /我的资源

	降序排序
	cloudpan189-go ls -desc 我的资源

	按文件大小降序排序
	cloudpan189-go ls -size -desc 我的资源
`,
			Category: "天翼云盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if config.Config.ActiveUser() == nil {
					fmt.Println("未登录账号")
					return nil
				}
				var (
					orderBy cloudpan.OrderBy = cloudpan.OrderByName
					orderSort cloudpan.OrderSort = cloudpan.OrderAsc
				)

				switch {
				case c.IsSet("asc"):
					orderSort = cloudpan.OrderAsc
				case c.IsSet("desc"):
					orderSort = cloudpan.OrderDesc
				default:
					orderSort = cloudpan.OrderAsc
				}

				switch {
				case c.IsSet("time"):
					orderBy = cloudpan.OrderByTime
				case c.IsSet("name"):
					orderBy = cloudpan.OrderByName
				case c.IsSet("size"):
					orderBy = cloudpan.OrderBySize
				default:
					orderBy = cloudpan.OrderByTime
				}

				command.RunLs(parseFamilyId(c), c.Args().Get(0), &command.LsOptions{
					Total: c.Bool("l") || c.Parent().Args().Get(0) == "ll",
				}, orderBy, orderSort)

				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "l",
					Usage: "详细显示",
				},
				cli.BoolFlag{
					Name:  "asc",
					Usage: "升序排序",
				},
				cli.BoolFlag{
					Name:  "desc",
					Usage: "降序排序",
				},
				cli.BoolFlag{
					Name:  "time",
					Usage: "根据时间排序",
				},
				cli.BoolFlag{
					Name:  "name",
					Usage: "根据文件名排序",
				},
				cli.BoolFlag{
					Name:  "size",
					Usage: "根据大小排序",
				},
				cli.StringFlag{
					Name:  "familyId",
					Usage: "家庭云ID",
					Value: "",
				},
			},
		},
		// 创建目录 mkdir
		{
			Name:      "mkdir",
			Usage:     "创建目录",
			UsageText: app.Name + " mkdir <目录>",
			Category:  "天翼云盘",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				if config.Config.ActiveUser() == nil {
					fmt.Println("未登录账号")
					return nil
				}
				command.RunMkdir(parseFamilyId(c), c.Args().Get(0))
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "familyId",
					Usage: "家庭云ID",
					Value: "",
				},
			},
		},
		// 删除文件/目录 rm
		{
			Name:      "rm",
			Usage:     "删除文件/目录",
			UsageText: app.Name + " rm <文件/目录的路径1> <文件/目录2> <文件/目录3> ...",
			Description: `
	注意: 删除多个文件和目录时, 请确保每一个文件和目录都存在, 否则删除操作会失败.
	被删除的文件或目录可在网盘文件回收站找回.

	示例:

	删除 /我的资源/1.mp4
	cloudpan189-go rm /我的资源/1.mp4

	删除 /我的资源/1.mp4 和 /我的资源/2.mp4
	cloudpan189-go rm /我的资源/1.mp4 /我的资源/2.mp4

	删除 /我的资源 整个目录 !!
	cloudpan189-go rm /我的资源
`,
			Category: "天翼云盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				if config.Config.ActiveUser() == nil {
					fmt.Println("未登录账号")
					return nil
				}
				command.RunRemove(parseFamilyId(c), c.Args()...)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "familyId",
					Usage: "家庭云ID",
					Value: "",
				},
			},
		},
		// 拷贝文件/目录 cp
		{
			Name:  "cp",
			Usage: "拷贝文件/目录",
			UsageText: app.Name + ` cp <文件/目录> <目标文件/目录>
	cloudpan189-go cp <文件/目录1> <文件/目录2> <文件/目录3> ... <目标目录>`,
			Description: `
	注意: 拷贝多个文件和目录时, 请确保每一个文件和目录都存在, 否则拷贝操作会失败.

	示例:

	将 /我的资源/1.mp4 复制到 根目录 /
	cloudpan189-go cp /我的资源/1.mp4 /

	将 /我的资源/1.mp4 和 /我的资源/2.mp4 复制到 根目录 /
	cloudpan189-go cp /我的资源/1.mp4 /我的资源/2.mp4 /
`,
			Category: "天翼云盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				if config.Config.ActiveUser() == nil {
					fmt.Println("未登录账号")
					return nil
				}
				if command.IsFamilyCloud(config.Config.ActiveUser().ActiveFamilyId) {
					fmt.Println("家庭云不支持复制操作")
					return nil
				}
				command.RunCopy(c.Args()...)
				return nil
			},
		},
		// 拷贝文件/目录到个人云/家庭云 xcp
		{
			Name:  "xcp",
			Usage: "拷贝个人云(家庭云)文件/目录到家庭云(个人云)",
			UsageText: app.Name + ` xcp <文件/目录>
	cloudpan189-go xcp <文件/目录1> <文件/目录2> <文件/目录3>`,
			Description: `
	注意: 拷贝多个文件和目录时, 请确保每一个文件和目录都存在, 否则拷贝操作会失败. 同样需要保证目标云不存在对应的文件，否则也会操作失败。

	示例:

	当前工作在个人云模式下，将 /个人云目录/1.mp4 复制到 家庭云根目录中
	cloudpan189-go xcp /个人云目录/1.mp4

	当前工作在家庭云模式下，将 /家庭云目录/1.mp4 和 /家庭云目录/2.mp4 复制到 个人云 /来自家庭共享 目录中
	cloudpan189-go xcp /家庭云目录/1.mp4 /家庭云目录/2.mp4
`,
			Category: "天翼云盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				if config.Config.ActiveUser() == nil {
					fmt.Println("未登录账号")
					return nil
				}
				familyId := parseFamilyId(c)
				fileSource := command.PersonCloud
				if c.IsSet("source") {
					sourceStr := c.String("source")
					if sourceStr == "person" {
						fileSource = command.PersonCloud
					} else if sourceStr == "family" {
						fileSource = command.FamilyCloud
					} else {
						fmt.Println("不支持的参数")
						return nil
					}
				} else {
					if command.IsFamilyCloud(config.Config.ActiveUser().ActiveFamilyId) {
						fileSource = command.FamilyCloud
					} else {
						fileSource = command.PersonCloud
					}
				}
				command.RunXCopy(fileSource, familyId, c.Args()...)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "familyId",
					Usage: "家庭云ID",
					Value: "",
					Required: false,
				},
				cli.StringFlag{
					Name:  "source",
					Usage: "文件源，person-个人云，family-家庭云",
					Value: "",
					Required: false,
				},
			},
		},
		// 移动文件/目录 mv
		{
			Name:  "mv",
			Usage: "移动文件/目录",
			UsageText: `移动:
	cloudpan189-go mv <文件/目录1> <文件/目录2> <文件/目录3> ... <目标目录>`,
			Description: `
	注意: 移动多个文件和目录时, 请确保每一个文件和目录都存在, 否则移动操作会失败.

	示例:

	将 /我的资源/1.mp4 移动到 根目录 /
	cloudpan189-go mv /我的资源/1.mp4 /
`,
			Category: "天翼云盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				if config.Config.ActiveUser() == nil {
					fmt.Println("未登录账号")
					return nil
				}

				command.RunMove(parseFamilyId(c), c.Args()...)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "familyId",
					Usage: "家庭云ID",
					Value: "",
				},
			},
		},
		// 重命名文件 rename
		{
			Name:  "rename",
			Usage: "重命名文件",
			UsageText: `重命名文件:
	cloudpan189-go rename <旧文件/目录名> <新文件/目录名>`,
			Description: `
	示例:

	将文件 1.mp4 重命名为 2.mp4
	cloudpan189-go rename 1.mp4 2.mp4

	将文件 /test/1.mp4 重命名为 /test/2.mp4
	要求必须是同一个文件目录内
	cloudpan189-go rename /test/1.mp4 /test/2.mp4
`,
			Category: "天翼云盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() != 2 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				if config.Config.ActiveUser() == nil {
					fmt.Println("未登录账号")
					return nil
				}
				command.RunRename(parseFamilyId(c), c.Args().Get(0), c.Args().Get(1))
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "familyId",
					Usage: "家庭云ID",
					Value: "",
				},
			},
		},
		// 分享文件/目录 share
		{
			Name:      "share",
			Usage:     "分享文件/目录",
			UsageText: app.Name + " share",
			Category:  "天翼云盘",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:        "set",
					Aliases:     []string{"s"},
					Usage:       "设置分享文件/目录",
					UsageText:   app.Name + " share set <文件/目录1> <文件/目录2> ...",
					Description: `
目前只支持创建私密链接.
示例:

    创建文件 1.mp4 的分享链接 
	cloudpan189-go share set 1.mp4

    创建文件 1.mp4 的分享链接，并指定有效期为1天
	cloudpan189-go share set -time 1 1.mp4
`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}
						if config.Config.ActiveUser() == nil {
							fmt.Println("未登录账号")
							return nil
						}
						if command.IsFamilyCloud(config.Config.ActiveUser().ActiveFamilyId) {
							fmt.Println("家庭云不支持文件分享，请切换到个人云")
							return nil
						}
						et := cloudpan.ShareExpiredTimeForever
						if c.IsSet("time") {
							op := c.String("time")
							if op == "1" {
								et = cloudpan.ShareExpiredTime1Day
							} else if op == "2" {
								et = cloudpan.ShareExpiredTime7Day
							} else {
								et = cloudpan.ShareExpiredTimeForever
							}
						}
						command.RunShareSet(c.Args(), et)
						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "time",
							Usage: "有效期，0-永久，1-1天，2-7天",
						},
					},
				},
				{
					Name:      "list",
					Aliases:   []string{"l"},
					Usage:     "列出已分享文件/目录",
					UsageText: app.Name + " share list",
					Action: func(c *cli.Context) error {
						command.RunShareList(c.Int("page"))
						return nil
					},
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "page",
							Usage: "分享列表的页数",
							Value: 1,
						},
					},
				},
				{
					Name:        "cancel",
					Aliases:     []string{"c"},
					Usage:       "取消分享文件/目录",
					UsageText:   app.Name + " share cancel <shareid_1> <shareid_2> ...",
					Description: `目前只支持通过分享id (shareid) 来取消分享.`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}
						command.RunShareCancel(converter.SliceStringToInt64(c.Args()))
						return nil
					},
				},
				{
					Name:        "save",
					Usage:       "转存分享的全部文件到指定文件夹",
					UsageText:   app.Name + " share save [save_dir_path] \"[share_url]\"",
					Description: `转存分享的全部文件到指定文件夹
示例:
    将 https://cloud.189.cn/t/RzUNre7nq2Uf 分享链接里面的全部文件转存到 /我的文档 这个网盘目录里面
    注意：转存需要一定的时间才能生效，需要等待一会才能完全转存到网盘文件夹里面
	cloudpan189-go share save /我的文档 https://cloud.189.cn/t/RzUNre7nq2Uf（访问码：io7x）
`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 2 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}
						if config.Config.ActiveUser() == nil {
							fmt.Println("未登录账号")
							return nil
						}
						command.RunShareSave(c.Args().Get(1), c.Args().Get(0))
						return nil
					},
				},
			},

		},
		// 上传文件/目录 upload
		{
			Name:      "upload",
			Aliases:   []string{"u"},
			Usage:     "上传文件/目录",
			UsageText: app.Name + " upload <本地文件/目录的路径1> <文件/目录2> <文件/目录3> ... <目标目录>",
			Description: `
	上传默认采用分片上传的方式, 上传的文件将会保存到, <目标目录>.

	示例:

	1. 将本地的 C:\Users\Administrator\Desktop\1.mp4 上传到网盘 /视频 目录
	注意区别反斜杠 "\" 和 斜杠 "/" !!!
	cloudpan189-go upload C:/Users/Administrator/Desktop/1.mp4 /视频

	2. 将本地的 C:\Users\Administrator\Desktop\1.mp4 和 C:\Users\Administrator\Desktop\2.mp4 上传到网盘 /视频 目录
	cloudpan189-go upload C:/Users/Administrator/Desktop/1.mp4 C:/Users/Administrator/Desktop/2.mp4 /视频

	3. 将本地的 C:\Users\Administrator\Desktop 整个目录上传到网盘 /视频 目录
	cloudpan189-go upload C:/Users/Administrator/Desktop /视频

	4. 使用相对路径
	cloudpan189-go upload 1.mp4 /视频

    5. 覆盖上传，已存在的同名文件会被移到回收站
	cloudpan189-go upload -ow 1.mp4 /视频
`,
			Category: "天翼云盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() < 2 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				subArgs := c.Args()
				command.RunUpload(subArgs[:c.NArg()-1], subArgs[c.NArg()-1], &command.UploadOptions{
					AllParallel:   c.Int("p"),
					Parallel:      1, // 天翼云盘一个文件只支持单线程上传
					MaxRetry:      c.Int("retry"),
					NoRapidUpload: c.Bool("norapid"),
					NoSplitFile:   true, // 天翼云盘不支持分片并发上传，只支持单线程上传，支持断点续传
					ShowProgress:  !c.Bool("np"),
					IsOverwrite:   c.Bool("ow"),
					FamilyId:      parseFamilyId(c),
				})
				return nil
			},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "p",
					Usage: "本次操作文件上传并发数量，即可以同时并发上传多少个文件。0代表跟从配置文件设置",
					Value: 0,
				},
				cli.IntFlag{
					Name:  "retry",
					Usage: "上传失败最大重试次数",
					Value: command.DefaultUploadMaxRetry,
				},
				cli.BoolFlag{
					Name:  "np",
					Usage: "no progress 不展示下载进度条",
				},
				cli.BoolFlag{
					Name:  "ow",
					Usage: "overwrite, 覆盖已存在的同名文件，注意已存在的文件会被移到回收站",
				},
				cli.BoolFlag{
					Name:  "norapid",
					Usage: "不检测秒传",
				},
				cli.StringFlag{
					Name:  "familyId",
					Usage: "家庭云ID",
					Value: "",
				},
			},
		},
		// 下载文件/目录 download
		{
			Name:      "download",
			Aliases:   []string{"d"},
			Usage:     "下载文件/目录",
			UsageText: app.Name + " download <文件/目录路径1> <文件/目录2> <文件/目录3> ...",
			Description: `
	下载的文件默认保存到, 程序所在目录的 download/ 目录.
	通过 cloudpan189-go config set -savedir <savedir>, 自定义保存的目录.
	支持多个文件或目录下载.
	自动跳过下载重名的文件!

	示例:

	设置保存目录, 保存到 D:\Downloads
	注意区别反斜杠 "\" 和 斜杠 "/" !!!
	cloudpan189-go config set -savedir D:\\Downloads
	或者
	cloudpan189-go config set -savedir D:/Downloads

	下载 /我的资源/1.mp4
	cloudpan189-go d /我的资源/1.mp4

	下载 /我的资源 整个目录!!
	cloudpan189-go d /我的资源

    下载 /我的资源/1.mp4 并保存下载的文件到本地的 d:/panfile
	cloudpan189-go d --saveto d:/panfile /我的资源/1.mp4
`,
			Category: "天翼云盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				// 处理saveTo
				var (
					saveTo string
				)
				if c.Bool("save") {
					saveTo = "."
				} else if c.String("saveto") != "" {
					saveTo = filepath.Clean(c.String("saveto"))
				}

				do := &command.DownloadOptions{
					IsPrintStatus:        c.Bool("status"),
					IsExecutedPermission: c.Bool("x"),
					IsOverwrite:          c.Bool("ow"),
					SaveTo:               saveTo,
					Parallel:             c.Int("p"),
					Load:                 c.Int("l"),
					MaxRetry:             c.Int("retry"),
					NoCheck:              c.Bool("nocheck"),
					ShowProgress:         !c.Bool("np"),
					FamilyId:             parseFamilyId(c),
				}

				command.RunDownload(c.Args(), do)
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "ow",
					Usage: "overwrite, 覆盖已存在的文件",
				},
				cli.BoolFlag{
					Name:  "status",
					Usage: "输出所有线程的工作状态",
				},
				cli.BoolFlag{
					Name:  "save",
					Usage: "将下载的文件直接保存到当前工作目录",
				},
				cli.StringFlag{
					Name:  "saveto",
					Usage: "将下载的文件直接保存到指定的目录",
				},
				cli.BoolFlag{
					Name:  "x",
					Usage: "为文件加上执行权限, (windows系统无效)",
				},
				cli.IntFlag{
					Name:  "p",
					Usage: "指定下载线程数",
				},
				cli.IntFlag{
					Name:  "l",
					Usage: "指定同时进行下载文件的数量",
				},
				cli.IntFlag{
					Name:  "retry",
					Usage: "下载失败最大重试次数",
					Value: pandownload.DefaultDownloadMaxRetry,
				},
				cli.BoolFlag{
					Name:  "nocheck",
					Usage: "下载文件完成后不校验文件",
				},
				cli.BoolFlag{
					Name:  "np",
					Usage: "no progress 不展示下载进度条",
				},
				cli.StringFlag{
					Name:  "familyId",
					Usage: "家庭云ID",
					Value: "",
				},
			},
		},
		// 回收站
		{
			Name:  "recycle",
			Usage: "回收站",
			Description: `
	回收站操作.

	示例:

	1. 从回收站还原两个文件, 其中的两个文件的 file_id 分别为 1013792297798440 和 643596340463870
	cloudpan189-go recycle restore 1013792297798440 643596340463870

	2. 从回收站删除两个文件, 其中的两个文件的 file_id 分别为 1013792297798440 和 643596340463870
	cloudpan189-go recycle delete 1013792297798440 643596340463870

	3. 清空回收站, 程序不会进行二次确认, 谨慎操作!!!
	cloudpan189-go recycle delete -all
`,
			Category: "天翼云盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NumFlags() <= 0 || c.NArg() <= 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
				}
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:      "list",
					Aliases:   []string{"ls", "l"},
					Usage:     "列出回收站文件列表",
					UsageText: app.Name + " recycle list",
					Action: func(c *cli.Context) error {
						command.RunRecycleList(c.Int("page"))
						return nil
					},
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "page",
							Usage: "回收站文件列表页数",
							Value: 1,
						},
					},
				},
				{
					Name:        "restore",
					Aliases:     []string{"r"},
					Usage:       "还原回收站文件或目录",
					UsageText:   app.Name + " recycle restore <file_id 1> <file_id 2> <file_id 3> ...",
					Description: `根据文件/目录的 fs_id, 还原回收站指定的文件或目录`,
					Action: func(c *cli.Context) error {
						if c.NArg() <= 0 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}
						command.RunRecycleRestore(c.Args()...)
						return nil
					},
				},
				{
					Name:        "delete",
					Aliases:     []string{"d"},
					Usage:       "删除回收站文件或目录 / 清空回收站",
					UsageText:   app.Name + " recycle delete [-all] <file_id 1> <file_id 2> <file_id 3> ...",
					Description: `根据文件/目录的 file_id 或 -all 参数, 删除回收站指定的文件或目录或清空回收站`,
					Action: func(c *cli.Context) error {
						if c.Bool("all") {
							// 清空回收站
							command.RunRecycleClear()
							return nil
						}

						if c.NArg() <= 0 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}
						command.RunRecycleDelete(c.Args()...)
						return nil
					},
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "all",
							Usage: "清空回收站, 程序不会进行二次确认, 谨慎操作!!!",
						},
					},
				},
			},
		},
		// 显示和修改程序配置项 config
		{
			Name:        "config",
			Usage:       "显示和修改程序配置项",
			Description: "显示和修改程序配置项",
			Category:    "配置",
			Before:      reloadFn,
			After:       saveFunc,
			Action: func(c *cli.Context) error {
				fmt.Printf("----\n运行 %s config set 可进行设置配置\n\n当前配置:\n", app.Name)
				config.Config.PrintTable()
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:      "set",
					Usage:     "修改程序配置项",
					UsageText: app.Name + " config set [arguments...]",
					Description: `
	注意:
		可通过设置环境变量 CLOUD189_CONFIG_DIR, 指定配置文件存放的目录.

		cache_size 的值支持可选设置单位, 单位不区分大小写, b 和 B 均表示字节的意思, 如 64KB, 1MB, 32kb, 65536b, 65536
		max_download_rate, max_upload_rate 的值支持可选设置单位, 单位为每秒的传输速率, 后缀'/s' 可省略, 如 2MB/s, 2MB, 2m, 2mb 均为一个意思

	例子:
		cloudpan189-go config set -cache_size 64KB
		cloudpan189-go config set -cache_size 16384 -max_download_parallel 200 -savedir D:/download`,
					Action: func(c *cli.Context) error {
						if c.NumFlags() <= 0 || c.NArg() > 0 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}
						if c.IsSet("cache_size") {
							err := config.Config.SetCacheSizeByStr(c.String("cache_size"))
							if err != nil {
								fmt.Printf("设置 cache_size 错误: %s\n", err)
								return nil
							}
						}
						if c.IsSet("max_download_parallel") {
							config.Config.MaxDownloadParallel = c.Int("max_download_parallel")
						}
						if c.IsSet("max_upload_parallel") {
							config.Config.MaxUploadParallel = c.Int("max_upload_parallel")
						}
						if c.IsSet("max_download_load") {
							config.Config.MaxDownloadLoad = c.Int("max_download_load")
						}
						if c.IsSet("max_download_rate") {
							err := config.Config.SetMaxDownloadRateByStr(c.String("max_download_rate"))
							if err != nil {
								fmt.Printf("设置 max_download_rate 错误: %s\n", err)
								return nil
							}
						}
						if c.IsSet("max_upload_rate") {
							err := config.Config.SetMaxUploadRateByStr(c.String("max_upload_rate"))
							if err != nil {
								fmt.Printf("设置 max_upload_rate 错误: %s\n", err)
								return nil
							}
						}
						if c.IsSet("savedir") {
							config.Config.SaveDir = c.String("savedir")
						}
						if c.IsSet("proxy") {
							config.Config.SetProxy(c.String("proxy"))
						}
						if c.IsSet("local_addrs") {
							config.Config.SetLocalAddrs(c.String("local_addrs"))
						}

						err := config.Config.Save()
						if err != nil {
							fmt.Println(err)
							return err
						}

						config.Config.PrintTable()
						fmt.Printf("\n保存配置成功!\n\n")

						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "cache_size",
							Usage: "下载缓存",
						},
						cli.IntFlag{
							Name:  "max_download_parallel",
							Usage: "下载网络连接的最大并发量",
						},
						cli.IntFlag{
							Name:  "max_upload_parallel",
							Usage: "上传网络连接的最大并发量",
						},
						cli.IntFlag{
							Name:  "max_download_load",
							Usage: "同时进行下载文件的最大数量",
						},
						cli.StringFlag{
							Name:  "max_download_rate",
							Usage: "限制最大下载速度, 0代表不限制",
						},
						cli.StringFlag{
							Name:  "max_upload_rate",
							Usage: "限制最大上传速度, 0代表不限制",
						},
						cli.StringFlag{
							Name:  "savedir",
							Usage: "下载文件的储存目录",
						},
						cli.StringFlag{
							Name:  "proxy",
							Usage: "设置代理, 支持 http/socks5 代理",
						},
						cli.StringFlag{
							Name:  "local_addrs",
							Usage: "设置本地网卡地址, 多个地址用逗号隔开",
						},
					},
				},
			},
		},
		// 清空控制台 clear
		{
			Name:        "clear",
			Aliases:     []string{"cls"},
			Usage:       "清空控制台",
			UsageText:   app.Name + " clear",
			Description: "清空控制台屏幕",
			Category:    "其他",
			Action: func(c *cli.Context) error {
				cmdliner.ClearScreen()
				return nil
			},
		},
		// 检测程序更新 update
		{
			Name:     "update",
			Usage:    "检测程序更新",
			Category: "其他",
			Action: func(c *cli.Context) error {
				if c.IsSet("y") {
					if !c.Bool("y") {
						return nil
					}
				}
				panupdate.CheckUpdate(app.Version, c.Bool("y"))
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "y",
					Usage: "确认更新",
				},
			},
		},
		// 退出程序 quit
		{
			Name:    "quit",
			Aliases: []string{"exit"},
			Usage:   "退出程序",
			Description: "退出程序",
			Category:    "其他",
			Action: func(c *cli.Context) error {
				return cli.NewExitError("", 0)
			},
			Hidden:   true,
			HideHelp: true,
		},
		// 调试用 debug
		{
			Name:        "debug",
			Aliases:     []string{"dg"},
			Usage:       "开发调试用",
			Description: "",
			Category:    "debug",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "param",
					Usage: "参数",
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	app.Run(os.Args)
}