# 关于
天翼云盘CLI，基于GO语言实现。仿 Linux shell 文件处理命令的天翼云盘命令行客户端。

# 特色
1. 多平台支持, 支持 Windows, macOS, linux, android, iOS等
2. 天翼云盘多用户支持
3. 支持个人云，家庭云无缝切换
4. 支持导入/导出功能，快速备份（导出）和恢复（导入）网盘文件。利用该功能可以进行跨网盘迁移文件
5. [下载](docs/manual.md#下载文件目录)网盘内文件, 支持多个文件或目录下载, 支持断点续传和单文件并行下载
6. [上传](docs/manual.md#上传文件目录)本地文件, 支持多个文件或目录上传

# 目录
- [关于](#关于)
- [特色](#特色)
- [如何安装](#如何安装)
  - [直接下载安装](#直接下载安装)
  - [apt安装](#apt安装)
  - [yum安装](#yum安装)
  - [brew安装](#brew安装)
  - [winget安装](#winget安装)
- [如何使用](#如何使用)
  - [基本使用](#基本使用)
    - [修改配置目录](#修改配置目录)
    - [启动程序](#启动程序)
    - [查看帮助](#查看帮助)
    - [登录](#登录)
    - [查看文件列表](#查看文件列表)
    - [下载文件](#下载文件)
    - [上传文件](#上传文件)
    - [创建分享链接](#创建分享链接)
    - [签到](#签到)
  - [更多命令](#更多命令)
- [常见问题](#常见问题)
  * [1. 如何开启Debug调试日志](#1-如何开启Debug调试日志)
- [交流反馈](#交流反馈)
- [鸣谢](#鸣谢)

# 如何安装
## 直接下载安装
可以直接在本仓库 [发布页](https://github.com/tickstep/cloudpan189-go/releases) 下载安装包，解压后使用。

要特别注意安装包的标签，不同的标签对应不同架构的系统，相关版本文件的标签说明如下：
1. arm / armv5 / armv7 : 适用32位ARM系统
2. arm64 : 适用64位ARM系统
3. 386 / x86 : 适用32系统，包括Intel和AMD的CPU系统
4. amd64 / x64 : 适用64位系统，包括Intel和AMD的CPU系统
5. mips : 适用MIPS指令集的CPU，例如国产龙芯CPU
6. macOS amd64适用Intel CPU的机器，macOS arm64目前主要是适用苹果M1芯片的机器
7. iOS arm64适用iPhone手机，并且必须是越狱的手机才能正常运行

参考例子：
```shell
wget https://github.com/tickstep/cloudpan189-go/releases/download/v0.1.3/cloudpan189-go-v0.1.3-linux-amd64.zip
unzip cloudpan189-go-v0.1.3-linux-amd64.zip
cd cloudpan189-go-v0.1.3-linux-amd64
./cloudpan189-go
```

## apt安装
适用于apt包管理器的系统，例如Ubuntu，国产deepin深度操作系统等。目前只支持amd64和arm64架构的机器。
```shell
sudo curl -fsSL http://file.tickstep.com/apt/pgp | gpg --dearmor | sudo tee /etc/apt/trusted.gpg.d/tickstep-packages-archive-keyring.gpg > /dev/null && echo "deb [signed-by=/etc/apt/trusted.gpg.d/tickstep-packages-archive-keyring.gpg arch=amd64,arm64] http://file.tickstep.com/apt cloudpan189-go main" | sudo tee /etc/apt/sources.list.d/tickstep-cloudpan189-go.list > /dev/null && sudo apt-get update && sudo apt-get install -y cloudpan189-go
 
```

## yum安装
适用于yum包管理器的系统，例如CentOS、RockyLinux等。目前只支持amd64和arm64架构的机器。
```shell
sudo curl -fsSL http://file.tickstep.com/rpm/cloudpan189-go/cloudpan189-go.repo | sudo tee /etc/yum.repos.d/tickstep-cloudpan189-go.repo > /dev/null && sudo yum install cloudpan189-go -y
 
```

## brew安装
适用于brew包管理器的系统，主要是苹果macOS系统。目前只支持amd64和arm64架构(Apple Silicon)的机器。
```shell
brew install cloudpan189-go
    
```
由于brew默认安装在系统目录下面，这样配置文件也默认存放在系统目录里了，建议设置系统变量进行配置文件的单独存储，例如
```shell
export CLOUD189_CONFIG_DIR=/Users/tickstep/Applications/cloud189/config
```

## winget安装
适用于Windows系统的winget包管理器。目前只支持x86和x64架构的机器。

更新源（可选）
```powershell
winget source update
 
```
安装
```powershell
winget install tickstep.cloudpan189-go --silent
 
```

# 如何使用
完整和详细的命令说明请查看手册：[命令手册](docs/manual.md)

1. Windows
   程序应在 命令提示符 (Command Prompt) 或 PowerShell 中运行.   
   也可直接双击程序运行, 具体使用方法请参见 [命令列表及说明](docs/manual.md#命令列表及说明)

2. Linux / macOS
   程序应在 终端 (Terminal) 运行.   
   具体使用方法请参见 [命令列表及说明](docs/manual.md#命令列表及说明)

如果程序运行时输出乱码, 请检查下终端的编码方式是否为 `UTF-8`.

如果没有带任何参数运行程序, 程序将会进入仿Linux shell系统用户界面的CLI交互模式, 可直接运行相关命令.   
在交互模式下, 光标所在行的前缀应为 `cloudpan189-go >`, 如果登录了帐号则格式为 `cloudpan189-go:<工作目录> <用户昵称>$ `

程序内置了相关命令的使用说明，你可以通过运行`命令 -h`的方式获取命令的使用说明，例如：`upload -h`获取上传命令的使用说明。

## 基本使用
本程序支持天翼云盘大多数命令操作，这里只介绍基本的使用，更多更详细的命令请查看手册：[命令手册](docs/manual.md)。

### 修改配置目录
你可以指定程序配置文件的存储路径，如果没有指定，程序会使用默认的目录。   
方法为设置环境变量`CLOUD189_CONFIG_DIR`并指定一个存在的目录，例如linux下面可以这样指定
```shell
export CLOUD189_CONFIG_DIR=/Users/tickstep/Applications/cloud189/config
```

### 启动程序
直接启动进入交互命令行
```shell
[tickstep@MacPro ~]$ cloudpan189-go
提示: 方向键上下可切换历史命令.
提示: Ctrl + A / E 跳转命令 首 / 尾.
提示: 输入 help 获取帮助.
cloudpan189-go > 
```

### 查看帮助
```shell
cloudpan189-go > help
...
   天翼云盘:
     backup           备份文件或目录
     cd               切换工作目录
     cp               拷贝文件/目录
     download, d      下载文件/目录
     export           导出文件/目录元数据
     import           导入文件
     ls, l, ll        列出目录
     mkdir            创建目录
     mv               移动文件/目录
     pwd              输出工作目录
     rapidupload, ru  手动秒传文件
     rename           重命名文件
     rm               删除文件/目录
     share            分享文件/目录
     upload, u        上传文件/目录
     xcp              转存拷贝文件/目录，个人云和家庭云之间转存文件
...
```

### 登录
需要先登录，已经登录过的可以跳过此步。登录成功后账号会加密存储在配置文件中，下一次程序启动会自动登录无需再次输入账号。
```shell
cloudpan189-go > login -username=131xxxxxx01@189.cn -password=123xxx

天翼云盘登录成功:  tickstep
cloudpan189-go:/ tickstep$ 
```

### 查看文件列表
```shell
cloudpan189-go:/ tickstep$ ls
  #   文件大小       修改日期               文件(目录)          
   0         -  2023-03-31 00:04:59  同步盘/                    
   1         -  2023-03-31 00:04:59  我的图片/                  
   2         -  2023-03-31 00:04:59  我的音乐/                  
   3         -  2023-03-31 00:04:59  我的视频/                  
   4         -  2023-03-31 00:04:59  我的文档/                  
   5         -  2023-03-31 00:04:59  我的应用/                        
   6         -  2022-09-07 18:43:00  我的项目/                  
   7         -  2023-03-26 22:28:07  cdn/                       
   8         -  2023-04-02 11:00:27  我的资源/                  
   9   47.83KB  2020-05-23 09:24:02  512.png                   
  10   55.35KB  2020-05-23 09:26:47  wx-mini-app-logo.png    
        总: 103.18KB                       文件总数: 2, 目录总数: 9
----
```

### 下载文件
```shell
cloudpan189-go:/ tickstep$ download 512.png

[0] 提示: 当前下载最大并发量为: 5, 下载缓存为: 65536
[1] 加入下载队列: /我的图片/512.png

[1] ----
文件ID: 5150329025489514
文件名: 512.png
文件类型: 文件
文件路径: /我的图片/512.png

[1] 准备下载: /我的图片/512.png
[1] 将会下载到路径: /Users/tickstep/Downloads/761169075/我的图片/512.png

[1] 下载开始


[1] 下载完成, 保存位置: /Users/tickstep/Downloads/761169075/我的图片/512.png
[1] 检验文件有效性成功: /Users/tickstep/Downloads/761169075/我的图片/512.png

下载结束, 时间: 1.361s, 数据总量: 47.830078KB
cloudpan189-go:/ tickstep$ 
```
支持并发同时下载文件，默认并发数是5个文件，你可以通过config进行修改。同时支持文件过滤。

### 上传文件
```shell
cloudpan189-go:/ tickstep$ upload /Users/tickstep/Downloads/app.zip /tmp
2023-04-06 22:06:17 [1] 加入上传队列: /Users/tickstep/Downloads/app.zip
[1] 准备上传: /Users/tickstep/Downloads/app.zip=>/tmp/app.zip
[1] 检测秒传中, 请稍候...
[1] 秒传失败，开始正常上传文件
[1] ↑ 7.21MB/7.21MB 0B/s in 3s ................
[1] 上传文件成功, 保存到网盘路径: /tmp/app.zip
2023-04-06 22:06:23 [1] 文件上传结果：成功！  耗时 5.400578399s

上传结束, 时间: 5.4s, 总大小: 7.211734MB
cloudpan189-go:/ tickstep$ 
```
上传也支持并发，默认并发数是10个文件，你可以通过config进行修改。同时支持文件过滤。

### 创建分享链接
```shell
cloudpan189-go:/ tickstep$ share set /tmp/app.zip 
路径: /tmp/app.zip
链接: https://cloud.189.cn/t/UbqQnuEnyAJn（访问码：nq0c）
cloudpan189-go:/ tickstep$ 
```

### 签到
```shell
cloudpan189-go:/ tickstep$ sign
签到成功，获得37M空间
第1次抽奖成功: 天翼云盘50M空间
第2次抽奖成功: 天翼云盘50M空间
cloudpan189-go:/ tickstep$ 
```

## 更多命令
更多更详细的命令请查看手册：[命令手册](docs/manual.md)。

# 常见问题
## 1 如何开启Debug调试日志
当需要定位问题，或者提交issue的时候抓取log，则需要开启debug日志。步骤如下：

### 第一步
Linux&MacOS   
命令行运行
```
export CLOUD189_VERBOSE=1
```

Windows   
不同版本会有些许不一样，请自行查询具体方法   
设置示意图如下：
![](./assets/images/win10-env-debug-config.png)

### 第二步
打开cloudpan189-go命令行程序，任何云盘命令都有类似如下日志输出
![](./assets/images/debug-log-screenshot.png)

# 交流反馈
提交issue: [issues页面](https://github.com/tickstep/cloudpan189-go/issues)   
联系邮箱: tickstep@outlook.com

# 鸣谢
本项目大量借鉴了以下相关项目的功能&成果   
> [iikira/BaiduPCS-Go](https://github.com/iikira/BaiduPCS-Go)   
> [Aruelius/cloud189](https://github.com/Aruelius/cloud189)   
