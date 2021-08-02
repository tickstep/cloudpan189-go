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
package main

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cmder"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/peterh/liner"
	"github.com/tickstep/cloudpan189-go/cmder/cmdliner"
	"github.com/tickstep/cloudpan189-go/cmder/cmdliner/args"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil/escaper"
	"github.com/tickstep/cloudpan189-go/internal/command"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/cloudpan189-go/internal/panupdate"
	"github.com/tickstep/cloudpan189-go/internal/utils"
	"github.com/tickstep/library-go/converter"
	"github.com/tickstep/library-go/logger"
	"github.com/urfave/cli"
)

const (
	// NameShortDisplayNum 文件名缩略显示长度
	NameShortDisplayNum = 16
)

var (
	// Version 版本号
	Version = "v0.1.0-dev"

	historyFilePath = filepath.Join(config.GetConfigDir(), "cloud189_command_history.txt")

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

func checkLoginExpiredAndRelogin() {
	cmder.ReloadConfigFunc(nil)
	activeUser := config.Config.ActiveUser()
	if activeUser == nil {
		// maybe expired, try to login
		cmder.TryLogin()
	}
	cmder.SaveConfigFunc(nil)
}

func main() {
	defer config.Config.Close()

	// check & relogin
	checkLoginExpiredAndRelogin()

	app := cli.NewApp()
	cmder.SetApp(app)

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

		os.Setenv(config.EnvVerbose, c.String("verbose"))
		isCli = true
		logger.Verbosef("提示: 你已经开启VERBOSE调试日志\n\n")

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
					"family", "export", "import", "backup",
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
				activeUser = config.Config.ActiveUser()
				runeFunc   = unicode.IsSpace
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
		cmder.ReloadConfigFunc(c)
		if config.Config.UpdateCheckInfo.LatestVer != "" {
			if utils.ParseVersionNum(config.Config.UpdateCheckInfo.LatestVer) > utils.ParseVersionNum(config.AppVersion) {
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
					logger.Verboseln("获取版本信息失败!")
					return
				}
				config.Config.UpdateCheckInfo.LatestVer = releaseInfo.TagName
				config.Config.UpdateCheckInfo.CheckTime = nowTime

				// save
				cmder.SaveConfigFunc(c)
			}
		}()

		for {
			var (
				prompt     string
				activeUser = config.Config.ActiveUser()
			)

			if activeUser == nil {
				activeUser = cmder.TryLogin()
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
		command.CmdLogin(),

		// 退出登录帐号 logout
		command.CmdLogout(),

		// 列出帐号列表 loglist
		command.CmdLoglist(),

		// 切换天翼帐号 su
		command.CmdSu(),

		// 获取当前帐号 who
		command.CmdWho(),

		// 切换家庭云 family
		command.CmdFamily(),

		// 获取当前帐号空间配额 quota
		command.CmdQuota(),

		// 用户签到 sign
		command.CmdSign(),

		// 切换工作目录 cd
		command.CmdCd(),

		// 输出工作目录 pwd
		command.CmdPwd(),

		// 列出目录 ls
		command.CmdLs(),

		// 创建目录 mkdir
		command.CmdMkdir(),

		// 删除文件/目录 rm
		command.CmdRm(),

		// 拷贝文件/目录 cp
		command.CmdCp(),

		// 拷贝文件/目录到个人云/家庭云 xcp
		command.CmdXcp(),

		// 移动文件/目录 mv
		command.CmdMv(),

		// 重命名文件 rename
		command.CmdRename(),

		// 分享文件/目录 share
		command.CmdShare(),

		// 备份 backup
		command.CmdBackup(),

		// 上传文件/目录 upload
		command.CmdUpload(),

		command.CmdRapidUpload(),

		// 下载文件/目录 download
		command.CmdDownload(),

		// 导出文件/目录元数据 export
		command.CmdExport(),

		// 导入文件 import
		command.CmdImport(),

		// 回收站
		command.CmdRecycle(),

		// 显示和修改程序配置项 config
		command.CmdConfig(),

		// 工具箱 tool
		command.CmdTool(),

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
			Name:        "quit",
			Aliases:     []string{"exit"},
			Usage:       "退出程序",
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
			Before:      cmder.ReloadConfigFunc,
			Action: func(c *cli.Context) error {
				os.Setenv(config.EnvVerbose, c.String("verbose"))
				fmt.Println("显示调试日志", logger.IsVerbose)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "param",
					Usage: "参数",
				},
				cli.BoolFlag{
					Name:        "verbose",
					Destination: &logger.IsVerbose,
					EnvVar:      config.EnvVerbose,
					Usage:       "显示调试信息",
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	app.Run(os.Args)
}
