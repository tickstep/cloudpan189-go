# 关于
天翼云盘CLI，基于GO语言实现。仿 Linux shell 文件处理命令的天翼云盘命令行客户端。

# 注意事项
*本项目还处于开发阶段，未经过充分的测试，如有bug欢迎提交issue.*   

# 编译
(待补充)

# 目录
- [特色](#特色)
- [下载/运行 说明](#下载/运行说明)
  * [Windows](#windows)
  * [Linux / macOS](#linux--macos)
- [命令列表及说明](#命令列表及说明)
  * [注意 ! ! !](#注意)
  * [登录天翼云盘帐号](#登录天翼云盘帐号)
  * [列出帐号列表](#列出帐号列表)
  * [获取当前帐号](#获取当前帐号)
  * [切换天翼云盘帐号](#切换天翼云盘帐号)
  * [退出天翼云盘帐号](#退出天翼云盘帐号)
  * [获取网盘配额](#获取网盘配额)
  * [切换工作目录](#切换工作目录)
  * [输出工作目录](#输出工作目录)
  * [列出目录](#列出目录)
  * [下载文件/目录](#下载文件/目录)
  * [上传文件/目录](#上传文件/目录)
  * [创建目录](#创建目录)
  * [删除文件/目录](#删除文件/目录)
  * [拷贝文件/目录](#拷贝文件/目录)
  * [移动文件/目录](#移动文件/目录)
  * [重命名文件/目录](#重命名文件/目录)
  * [分享文件/目录](#分享文件/目录)
    + [设置分享文件/目录](#设置分享文件/目录)
    + [列出已分享文件/目录](#列出已分享文件/目录)
    + [取消分享文件/目录](#取消分享文件/目录)
  * [显示和修改程序配置项](#显示和修改程序配置项)
- [初级使用教程](#初级使用教程)
  * [1. 查看程序使用说明](#1-查看程序使用说明)
  * [2. 登录天翼云盘帐号 (必做)](#2-登录帐号)
  * [3. 切换网盘工作目录](#3-切换网盘工作目录)
  * [4. 网盘内列出文件和目录](#4-网盘内列出文件和目录)
  * [5. 下载文件](#5-下载文件)
  * [6. 设置下载最大并发量](#6-设置下载最大并发量)
  * [7. 退出程序](#7-退出程序)
- [交流反馈](#交流反馈)
- [鸣谢](#鸣谢)


# 特色
多平台支持, 支持 Windows, macOS, linux等.

天翼云盘多用户支持;

[下载](#下载文件/目录)网盘内文件, 支持多个文件或目录下载, 支持断点续传和单文件并行下载;

[上传](#上传文件/目录)本地文件, 支持多个文件或目录上传;



# 下载/运行说明

可以直接在[发布页](https://github.com/tickstep/cloudpan189-go/releases)下载使用.

如果程序运行时输出乱码, 请检查下终端的编码方式是否为 `UTF-8`.

使用本程序之前, 非常建议先学习一些 linux 基础命令知识.

如果没有带任何参数运行程序, 程序将会进入仿Linux shell系统用户界面的cli交互模式, 可直接运行相关命令.

cli交互模式下, 光标所在行的前缀应为 `cloudpan189-go >`, 如果登录了帐号则格式为 `cloudpan189-go:<工作目录> <用户ID>$ `

程序会提供相关命令的使用说明.

## Windows

程序应在 命令提示符 (Command Prompt) 或 PowerShell 中运行.

也可直接双击程序运行, 具体使用方法请参见 [命令列表及说明](#命令列表及说明) 和 [初级使用教程](#初级使用教程).

## Linux / macOS

程序应在 终端 (Terminal) 运行.

具体使用方法请参见 [命令列表及说明](#命令列表及说明) 和 [初级使用教程](#初级使用教程).


# 命令列表及说明

## 注意

命令的前缀 `cloudpan189-go` 为指向程序运行的全路径名 (ARGv 的第一个参数)

直接运行程序时, 未带任何其他参数, 则程序进入cli交互模式, 进入cli模式运行以下命令时要把命令的前缀 `cloudpan189-go` 去掉! 即不需要输入`cloudpan189-go`。

cli交互模式已支持按tab键自动补全命令.

## 登录天翼云盘帐号

### 直接登录

```
cloudpan189-go login
```
### 例子
```
cloudpan189-go login
请输入用户名(手机号/邮箱/别名), 回车键提交 > 1234567
```


## 列出帐号列表

```
cloudpan189-go loglist
```

列出所有已登录的帐号

## 获取当前帐号

```
cloudpan189-go who
```

## 切换天翼云盘帐号

切换已登录的帐号
```
cloudpan189-go su <uid>
```
```
cloudpan189-go su

请输入要切换帐号的 # 值 >
```

## 退出天翼云盘帐号

退出当前登录的帐号
```
cloudpan189-go logout
```

程序会进一步确认退出帐号, 防止误操作.

## 获取网盘配额

```
cloudpan189-go quota
```
获取网盘的总储存空间, 和已使用的储存空间

## 切换工作目录
```
cloudpan189-go cd <目录>
```

### 例子
```
# 切换 /我的文档 工作目录
cloudpan189-go cd /我的文档

# 切换 上级目录
cloudpan189-go cd ..

# 切换 根目录
cloudpan189-go cd /

```

## 输出工作目录
```
cloudpan189-go pwd
```

## 列出目录

列出当前工作目录的文件和目录或指定目录
```
cloudpan189-go ls
```
```
cloudpan189-go ls <目录>
```

### 可选参数
```
-asc: 升序排序
-desc: 降序排序
-time: 根据时间排序
-name: 根据文件名排序
-size: 根据大小排序
```

### 例子
```
# 列出 我的文档 内的文件和目录
cloudpan189-go ls 我的文档

# 绝对路径
cloudpan189-go ls /我的文档

# 降序排序
cloudpan189-go ls -desc 我的文档

# 按文件大小降序排序
cloudpan189-go ls -size -desc 我的文档
```

## 下载文件/目录
```
cloudpan189-go download <网盘文件或目录的路径1> <文件或目录2> <文件或目录3> ...
cloudpan189-go d <网盘文件或目录的路径1> <文件或目录2> <文件或目录3> ...
```

### 可选参数
```
  --ow            overwrite, 覆盖已存在的文件
  --status        输出所有线程的工作状态
  --save          将下载的文件直接保存到当前工作目录
  --saveto value  将下载的文件直接保存到指定的目录
  -x              为文件加上执行权限, (windows系统无效)
  -p value        指定下载线程数 (default: 0)
  -l value        指定同时进行下载文件的数量 (default: 0)
  --retry value   下载失败最大重试次数 (default: 3)
  --nocheck       下载文件完成后不校验文件
```


### 例子
```
# 设置保存目录, 保存到 D:\Downloads
# 注意区别反斜杠 "\" 和 斜杠 "/" !!!
cloudpan189-go config set -savedir D:/Downloads

# 下载 /我的文档/1.mp4
cloudpan189-go d /我的文档/1.mp4

# 下载 /我的文档 整个目录!!
cloudpan189-go d /我的文档
```

下载的文件默认保存到 **程序所在目录** 的 download/ 目录, 支持设置指定目录, 重名的文件会自动跳过!

通过 `cloudpan189-go config set -savedir <savedir>` 可以自定义保存的目录.

支持多个文件或目录下载.

自动跳过下载重名的文件!

## 上传文件/目录
```
cloudpan189-go upload <本地文件/目录的路径1> <文件/目录2> <文件/目录3> ... <目标目录>
cloudpan189-go u <本地文件/目录的路径1> <文件/目录2> <文件/目录3> ... <目标目录>
```

### 例子:
```
# 将本地的 C:\Users\Administrator\Desktop\1.mp4 上传到网盘 /视频 目录
# 注意区别反斜杠 "\" 和 斜杠 "/" !!!
cloudpan189-go upload C:/Users/Administrator/Desktop/1.mp4 /视频

# 将本地的 C:\Users\Administrator\Desktop\1.mp4 和 C:\Users\Administrator\Desktop\2.mp4 上传到网盘 /视频 目录
cloudpan189-go upload C:/Users/Administrator/Desktop/1.mp4 C:/Users/Administrator/Desktop/2.mp4 /视频

# 将本地的 C:\Users\Administrator\Desktop 整个目录上传到网盘 /视频 目录
cloudpan189-go upload C:/Users/Administrator/Desktop /视频
```

## 创建目录
```
cloudpan189-go mkdir <目录>
```

### 例子
```
cloudpan189-go mkdir test123
```

## 删除文件/目录
```
cloudpan189-go rm <网盘文件或目录的路径1> <文件或目录2> <文件或目录3> ...
```

注意: 删除多个文件和目录时, 请确保每一个文件和目录都存在, 否则删除操作会失败.

被删除的文件或目录可在网盘文件回收站找回.

### 例子
```
# 删除 /我的文档/1.mp4
cloudpan189-go rm /我的文档/1.mp4

# 删除 /我的文档/1.mp4 和 /我的文档/2.mp4
cloudpan189-go rm /我的文档/1.mp4 /我的文档/2.mp4

# 删除 /我的文档 整个目录 !!
cloudpan189-go rm /我的文档
```


## 拷贝文件/目录
```
cloudpan189-go cp <文件/目录> <目标 文件/目录>
cloudpan189-go cp <文件/目录1> <文件/目录2> <文件/目录3> ... <目标目录>
```

注意: 拷贝多个文件和目录时, 请确保每一个文件和目录都存在, 否则拷贝操作会失败.

### 例子
```
# 将 /我的文档/1.mp4 复制到 根目录 /
cloudpan189-go cp /我的文档/1.mp4 /

# 将 /我的文档/1.mp4 和 /我的文档/2.mp4 复制到 根目录 /
cloudpan189-go cp /我的文档/1.mp4 /我的文档/2.mp4 /
```

## 移动文件/目录
```
cloudpan189-go mv <文件/目录1> <文件/目录2> <文件/目录3> ... <目标目录>
```

注意: 移动多个文件和目录时, 请确保每一个文件和目录都存在, 否则移动操作会失败.

### 例子
```
# 将 /我的文档/1.mp4 移动到 根目录 /
cloudpan189-go mv /我的文档/1.mp4 /
```

## 重命名文件/目录
```
cloudpan189-go rename <旧文件/目录名> <新文件/目录名>
```

注意: 重命名的文件/目录，如果指定的是绝对路径，则必须保证新旧的绝对路径在同一个文件夹内，否则重命名失败！

### 例子
```
# 将 /我的文档/1.mp4 重命名为 /我的文档/2.mp4
cloudpan189-go rename /我的文档/1.mp4 /我的文档/2.mp4
```

## 分享文件/目录
```
cloudpan189-go share
```

### 设置分享文件/目录
```
cloudpan189-go share set <文件/目录1> <文件/目录2> ...
cloudpan189-go share s <文件/目录1> <文件/目录2> ...
```

### 列出已分享文件/目录
```
cloudpan189-go share list
cloudpan189-go share l
```

### 取消分享文件/目录
```
cloudpan189-go share cancel <shareid_1> <shareid_2> ...
cloudpan189-go share c <shareid_1> <shareid_2> ...
```

目前只支持通过分享id (shareid) 来取消分享.

## 显示和修改程序配置项
```
# 显示配置
cloudpan189-go config

# 设置配置
cloudpan189-go config set
```


### 例子
```
# 显示所有可以设置的值
cloudpan189-go config -h
cloudpan189-go config set -h

# 设置下载文件的储存目录
cloudpan189-go config set -savedir D:/Downloads

# 设置下载最大并发量为 150
cloudpan189-go config set -max_download_parallel 15

# 组合设置
cloudpan189-go config set -max_download_parallel 15 -savedir D:/Downloads
```


# 初级使用教程

新手建议: **双击运行程序**, 进入仿 Linux shell 的 cli 交互模式;

cli交互模式下, 光标所在行的前缀应为 `cloudpan189-go >`, 如果登录了帐号则格式为 `cloudpan189-go:<工作目录> <用户ID>$ `

以下例子的命令, 均为 cli交互模式下的命令

运行命令的正确操作: **输入命令, 按一下回车键 (键盘上的 Enter 键)**, 程序会接收到命令并输出结果

## 1 查看程序使用说明

cli交互模式下, 运行命令 `help`

## 2 登录帐号
*以下所有操作都必须在登录账户后才能进行*

cli交互模式下, 运行命令 `login -h` (注意空格) 查看帮助

cli交互模式下, 运行命令 `login` 程序将会提示你输入用户名(手机号/邮箱/别名)和密码

## 3 切换网盘工作目录

cli交互模式下, 运行命令 `cd /我的文档` 将工作目录切换为 `/我的文档` (前提: 该目录存在于网盘)

将工作目录切换为 `/我的文档` 成功后, 运行命令 `cd ..` 切换上级目录, 即将工作目录切换为 `/`

## 4 网盘内列出文件和目录

cli交互模式下, 运行命令 `ls -h` (注意空格) 查看帮助

cli交互模式下, 运行命令 `ls` 来列出当前所在目录的文件和目录

cli交互模式下, 运行命令 `ls /我的文档` 来列出 `/我的文档` 内的文件和目录

## 5 下载文件

说明: 下载的文件默认保存到 download/ 目录 (文件夹)

cli交互模式下, 运行命令 `d -h` (注意空格) 查看帮助

cli交互模式下, 运行命令 `d /我的文档/1.mp4` 来下载位于 `/我的文档/1.mp4` 的文件 `1.mp4` , 该操作等效于运行以下命令:

```
cd /我的文档
d 1.mp4
```

支持目录 (文件夹) 下载, 所以, 运行以下命令, 会下载 `/我的文档` 内的所有文件:

```
d /我的文档
```

## 6 设置下载最大并发量

cli交互模式下, 运行命令 `config set -h` (注意空格) 查看设置帮助以及可供设置的值

cli交互模式下, 运行命令 `config set -max_download_parallel 2` 将下载最大并发量设置为 2

注意：下载最大并发量的值不易设置过高, 可能会导致被限速

## 7 退出程序

运行命令 `quit` 或 `exit` 或 组合键 `Ctrl+C`


# 交流反馈

提交issue: [issues页面](https://github.com/tickstep/cloudpan189-go/issues)   
联系邮箱: tickstep@outlook.com

# 鸣谢
本项目大量借鉴了以下相关项目的功能&成果   
> [iikira/cloudpan189-go](https://github.com/iikira/cloudpan189-go)   
> [Aruelius/cloud189](https://github.com/Aruelius/cloud189)   
