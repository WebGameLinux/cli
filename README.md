# cli

#### 介绍
golang cli 工具包

#### 软件架构
golang 命令行 应用构建 工具包
在 github.com/urfave/cli/v2 基础扩展与简化


#### 安装教程
```bash
 $ go get github.com/WebGameLinux/cli
```

#### 使用说明

eg :+1 

command.go
```go
package examples 
import (
		"fmt"
		"github.com/WebGameLinux/cli/core"
		"github.com/WebGameLinux/cli/resolver"
		"github.com/urfave/cli/v2"
)
// 支持字符串命令参数表达式 解析 
// @ : required 
// # : hidden 
// -- flag的 name
// -x 表单对应flag 的别名
// 每个flag 表达式 有3 部分
//  name : usage : default-value 
// name 部分包括 了 flag 类型定义 和 是否必选参标识 
// 参数类型 使用 中括号包括 "[xxx]"
// 参数类型 支持 string, number, float ,bool ,intArr 
// 每个 flag 使用换行符分隔 

const (
		VArgs = `
@[string]--config,-c: 配置文件路径 : file.txt
[bool]--on :  是否开启 : false 
--version,-V,-ver : 显示版本号 : 0.0.1
`
		VName        = "version:test"
		VUsage       = "输出版本帮助"
		VDescription = "描述"
)

type VersionCommand struct {
		core.BaseCommander
}

func NewVersionCmd() *VersionCommand {
		return &VersionCommand{}
}

// 自动 载入相关信息
func (cmd *VersionCommand) Boot() {
		cmd.Command.Name = VName
		cmd.Command.Usage = VUsage
		cmd.Command.Description = VDescription
		flags := resolver.StrTemplateResolver(VArgs).Args()
		cmd.Command.Flags = flags
}

// 必须 实现 自动注册
func (cmd *VersionCommand) Register(app *core.CmdApp) {
		core.CommanderRegisterFn(cmd, app)
}

// 命令处理
func (cmd *VersionCommand) Handle(ctx *cli.Context) error {
		fmt.Printf("config:%s \n", ctx.String("config"))
		fmt.Printf("on:%v \n", ctx.Bool("on"))
		fmt.Printf("version:%v \n", ctx.String("version"))
		return nil
}
```

main.go
```go
package main
import (
		"fmt"
		"github.com/WebGameLinux/cli/core"
		"github.com/WebGameLinux/cli/examples"
		"github.com/urfave/cli/v2"
		"log"
		"os"
)

func main() {
		app := core.NewCmdApp()
		cmd := examples.NewVersionCmd()
		cmd.Register(app)
                // 启动命令应用
		app.Run()
}
```

examples目录中有相关案例代码

#### 其他功能

简单 命令 flag 封装 与 解析器 
```go
package examples

import (
"fmt"
"github.com/WebGameLinux/cli/resolver"
)

const (
		FlagsExpress = `
[bool]--help : show help menus : false
--conf : config file path  : ./config.ini
--name : set process name  :  cli
[bool]--deamon : process run in background : true
`
)
// SimpleCommanderApp 适用于需要简单命令行嵌入支持的应用
 
func GetCommander() *resolver.SimpleCommanderApp {
		// 设置启动时 需要注册 命令行参数 即可
		resolver.SimpleCommandLine.BootHandler = func(app *resolver.SimpleCommanderApp) {
			// 支持字符串表达解析注册 (批量)
            app.AddFlagsByStrExpress(FlagsExpress)
            // 支持单条注册 [支持类型: string,int,float64,bool]
            app.AddFlag("string", "test","命令参数", "默认值")
		}
		return resolver.SimpleCommandLine
}

func GetArg() {
		app := GetCommander()
		app.Run()
		if app.GetBool("help") {
				app.PrintUsage()
		} else {
				fmt.Println(app.GetBool("deamon"))
				fmt.Println(app.Get("conf"))
				fmt.Println(app.Get("name"))
				fmt.Println(app.Get("test"))
		}
}

```

以守护进程方式启动 或 杀死对应守护进程
```go
package main

import (
		_ "github.com/WebGameLinux/cli/daemon"
		"log"
		"net/http"
)

func main() {
		mux := http.NewServeMux()
		mux.HandleFunc("/index", func(rw http.ResponseWriter, req *http.Request) {
				rw.Write([]byte("hello, golang!\n"))
		})
		log.Fatalln(http.ListenAndServe(":7070", mux))
}
```

使用方式 : 
```bash
$ go build -o web
$ ./web -d=true --name=web # 后台启动 
$ ./web --stop --name=web  # 关闭进程 
```
支持 pidFile save  
支持 .daemon.ini 文件配置  
其中 ```--name``` 为指定应用名,  
也指定了应用 读取了 配置文件的作用域  
pidFile 配置 在 .daemon.ini 中  

```ini
[web]
deamonOn=true
;; pid保存的文件路径
pidFile=web.pid
```

pidFile 也可在环境变量中指定(配置文件优先)
```bash
$ export Daemon.PidFile="web.pid"
$ ./web -d=true --name=web 
```

#### windows 启动后台常驻进程 plugin

为兼容 daemon 后台常驻功能 需要使用windows 后台服务注册

plugin 目录下win 中的 WinSw.exe 就是 windows 服务注册的工具

编辑 相关 配置 WinSw.xml 为对应 exe 即可

```xml
<service>
 <!--可执行二进制文件-->
    <executable>%BASE%\web.exe</executable>
<!--启动相关参数-->
    <arguments></arguments>
</service>
```

将 WinSW.exe 修改程 对应 应用名
配置也改程对应名 
eg : 
```cmd
rename  WinSw.exe Web-cli.exe 
cp  WinSw.xml Web-cli.xml 
```
然后 将cli、(xml)配置 和 可以执行文件 拷贝 到同一个文件目录下

```cmd
:: 第一次,注册后台常驻进程服务 (install一次即可)
Web-cli.exe install  
:: 启动服务
Web-cli.exe start    
:: 停止服务
Web-cli.exe stop     
```


#### 软件依赖 
感谢🙏  开源工具包 urfave/cli/v2 
感谢🙏  window开源工具包 winsw

```
    github.com/mitchellh/mapstructure v1.2.2
    github.com/urfave/cli/v2          v2.2.0
```
[mapstructure](https://github.com/mitchellh/mapstructure)   
[cli](https://github.com/urfave/cli)   
[winsw](https://github.com/winsw/winsw)