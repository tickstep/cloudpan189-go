package main

import (
	"fmt"
	"github.com/peterh/liner"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdliner"
	"github.com/tickstep/cloudpan189-go/cmder/cmdliner/args"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil/converter"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil/escaper"
	"github.com/tickstep/cloudpan189-go/internal/command"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"github.com/urfave/cli"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
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
					"clear", "quit", "exit", "quota", "who",
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
				cookieOfToken := ""
				if c.IsSet("COOKIE_LOGIN_USER") {
					cookieOfToken = c.String("COOKIE_LOGIN_USER")
				} else if c.NArg() == 0 {
					var err error
					cookieOfToken, err = command.RunLogin(c.String("username"), c.String("password"))
					if err != nil {
						fmt.Println(err)
						return err
					}
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				cloudUser, _ := config.SetupUserByCookie(cookieOfToken)
				config.Config.SetActiveUser(cloudUser)
				fmt.Println("天翼帐号登录成功: ", cloudUser.Nickname)
				return nil
			},
			// 命令的附加options参数说明，使用 help login 命令即可查看
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "username",
					Usage: "登录百度帐号的用户名(手机号/邮箱/用户名)",
				},
				cli.StringFlag{
					Name:  "password",
					Usage: "登录百度帐号的用户名的密码",
				},
				cli.StringFlag{
					Name:  "COOKIE_LOGIN_USER",
					Usage: "使用 COOKIE_LOGIN_USER cookie来登录帐号",
				},
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
				r, _ := config.Config.ActiveUser().PanClient().FileSearch(cloudpan.NewFileSearchParam())
				fmt.Printf("%+v", r)
				return nil
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	app.Run(os.Args)
}