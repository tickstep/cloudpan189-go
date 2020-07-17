package main

import (
	"fmt"
	"github.com/peterh/liner"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdliner"
	"github.com/tickstep/cloudpan189-go/cmder/cmdliner/args"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil/escaper"
	"github.com/tickstep/cloudpan189-go/internal/command"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/cloudpan189-go/library/converter"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"github.com/urfave/cli"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const (
	// NameShortDisplayNum 文件名缩略显示长度
	NameShortDisplayNum = 16
)

var (
	// Version 版本号
	Version = "v1.0.0-dev"

	historyFilePath = filepath.Join(config.GetConfigDir(), "cloud189_command_history.txt")
	reloadFn        = func(c *cli.Context) error {
		err := config.Config.Reload()
		if err != nil {
			fmt.Printf("重载配置错误: %s\n", err)
		}
		return nil
	}
	saveFunc = func(c *cli.Context) error {
		err := config.Config.Save()
		if err != nil {
			fmt.Printf("保存配置错误: %s\n", err)
		}
		return nil
	}

	isCli bool
)

func init() {
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

func main()  {
	defer config.Config.Close()

	app := cli.NewApp()
	app.Name = "cloudpan189-go"
	app.Version = Version
	app.Author = "tickstep/cloudpan189-go: https://github.com/tickstep/cloudpan189-go"
	app.Copyright = "(c) 2020 tickstep."
	app.Usage = "天翼云盘客户端 for " + runtime.GOOS + "/" + runtime.GOARCH
	app.Description = `cloudpan189-go 使用Go语言编写的天翼云盘命令行客户端, 为操作天翼云盘, 提供实用功能.
	具体功能, 参见 COMMANDS 列表

	---------------------------------------------------
	前往 https://github.com/tickstep/cloudpan189-go 以获取更多帮助信息!
	前往 https://github.com/tickstep/cloudpan189-go/releases 以获取程序更新信息!
	---------------------------------------------------

	交流反馈:
		提交Issue: https://github.com/tickstep/cloudpan189-go/issues
		邮箱: tickstep@outlook.com`

	// 全局options
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "启用调试",
			EnvVar:      logger.EnvVerbose,
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
					"cd", "cp", "download", "ls", "mkdir", "mv", "rm", "share", "upload", "login",
					"clear", "quit", "exit", "quota", "who", "sign",
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

		for {
			var (
				prompt     string
				activeUser = config.Config.ActiveUser()
			)

			if activeUser != nil && activeUser.Nickname != "" {
				// 格式: cloudpan189-go:<工作目录> <UserName>$
				// 工作目录太长时, 会自动缩略
				prompt = app.Name + ":" + converter.ShortDisplay(path.Base(activeUser.Workdir), NameShortDisplayNum) + " " + activeUser.Nickname + "$ "
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
		cloudpan189-go login -username=tickstep
        cloudpan189-go login -COOKIE_LOGIN_USER=8B12CBBCE89CA8DFC3445985B63B511B5E7EC7...

	常规登录:
		按提示一步一步来即可.
`,
			Category: "天翼云盘账号",
			Before:   reloadFn, // 每次进行登录动作的时候需要调用刷新配置
			After:    saveFunc, // 登录完成需要调用保存配置
			Action: func(c *cli.Context) error {
				appToken := cloudpan.AppLoginToken{}
				webToken := cloudpan.WebLoginToken{}
				if c.IsSet("COOKIE_LOGIN_USER") {
					webToken.CookieLoginUser = c.String("COOKIE_LOGIN_USER")
				} else if c.NArg() == 0 {
					var err error
					webToken, appToken, err = command.RunLogin(c.String("username"), c.String("password"))
					if err != nil {
						fmt.Println(err)
						return err
					}
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				cloudUser, _ := config.SetupUserByCookie(webToken, appToken)
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
					Usage: "登录天翼帐号的用户名的密码",
				},
				cli.StringFlag{
					Name:  "COOKIE_LOGIN_USER",
					Usage: "使用 COOKIE_LOGIN_USER cookie来登录帐号",
				},
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

				fmt.Printf("切换用户: %s\n", switchedUser.Nickname)
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
				fmt.Printf("当前帐号 uid: %d, 昵称: %s, 用户名: %s, 性别: %s\n", activeUser.UID, activeUser.Nickname, activeUser.AccountName, gender)
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
					fmt.Printf("账号: %s, 个人空间总额 uid: %s, 个人空间已使用: %s, 比率: %f%%\n",
						config.Config.ActiveUser().Nickname,
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
				command.RunChangeDirectory(c.Args().Get(0))
				return nil
			},
			Flags: []cli.Flag{
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
				fmt.Println(config.Config.ActiveUser().Workdir)
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

				command.RunLs(c.Args().Get(0), &command.LsOptions{
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
				command.RunMkdir(c.Args().Get(0))
				return nil
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
				command.RunRemove(c.Args()...)
				return nil
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
				command.RunCopy(c.Args()...)
				return nil
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

				command.RunMove(c.Args()...)
				return nil
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
				command.RunRename(c.Args().Get(0), c.Args().Get(1))
				return nil
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
	遇到同名文件将会自动覆盖!!

	禁用分片上传时只能使用单线程上传, 指定的单个文件上传最大线程数将会无效.

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
					AllParallel:   c.Int("all_parallel"),
					Parallel:      1, // 天翼云盘只支持单线程上传
					MaxRetry:      c.Int("retry"),
					NoRapidUpload: c.Bool("norapid"),
					NoSplitFile:   true, // 天翼云盘不支持分片并发上传，只支持单线程上传，支持断点续传
				})
				return nil
			},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "all_parallel",
					Usage: "所有文件并发上传数量，即可以同时并发上传多少个文件",
					Value: command.DefaultUploadMaxAllParallel,
				},
				cli.IntFlag{
					Name:  "retry",
					Usage: "上传失败最大重试次数",
					Value: command.DefaultUploadMaxRetry,
				},
				cli.BoolFlag{
					Name:  "norapid",
					Usage: "不检测秒传",
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
				//activeUser := config.Config.ActiveUser()
				//uploadFile := "/Volumes/Downloads/tmp/pcs_config.json"
				//uploadFileData,_ := ioutil.ReadFile(uploadFile)
				//p := &cloudpan.AppCreateUploadFileParam{
				//	ParentFolderId: "21491110455851923",
				//	FileName: "pcs_config.json",
				//	Size: int64(len(uploadFileData)),
				//	Md5: hash.Md5OfBytes(uploadFileData),
				//	LastWrite: "2013-06-16 24:35:08",
				//	LocalPath: uploadFile,
				//}
				//fmt.Println("create upload file")
				//r1, err := config.Config.ActiveUser().PanClient().AppCreateUploadFile(p)
				//if err != nil {
				//	fmt.Println(err)
				//	return nil
				//}
				//fmt.Println("upload file data")
				//err = config.Config.ActiveUser().PanClient().AppUploadFileData(r1.FileUploadUrl, r1.UploadFileId, r1.XRequestId,
				//	&cloudpan.AppFileRange{0, len(uploadFileData)}, uploadFileData)
				//if err != nil {
				//	fmt.Println(err)
				//	return nil
				//}
				//fmt.Println("commit")
				//r2, err := config.Config.ActiveUser().PanClient().AppUploadFileCommit(r1.FileCommitUrl, r1.UploadFileId, r1.XRequestId)
				//if err != nil {
				//	fmt.Println(err)
				//	return nil
				//}
				//fmt.Printf("%+v", r2)
				//durl, e := activeUser.PanClient().AppGetFileDownloadUrl("21301210456083931")
				//if e != nil {
				//	fmt.Println(e)
				//	return nil
				//}
				//fmt.Println(durl)
				//resp, e := activeUser.PanClient().AppDownloadFileData(durl, cloudpan.AppFileRange{4372546,14372545})
				//if e != nil {
				//	fmt.Println(e)
				//	return nil
				//}
				//f, er := os.OpenFile("D:/dl/189test.pdf", os.O_APPEND | os.O_CREATE, 0777)
				//if er != nil {
				//	fmt.Println(e)
				//	return nil
				//}
				//defer f.Close()
				//resp.Write(f)
				//r,e := activeUser.PanClient().AppGetFileInfo(&cloudpan.AppGetFileInfoParam{
				//	"", "/Cloud189-Go/tup/image_tool1.bat",
				//})
				//if e != nil {
				//	fmt.Println(e)
				//	return nil
				//}
				//fmt.Printf("%+v", r)
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