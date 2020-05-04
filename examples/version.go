package examples

import (
		"fmt"
		"github.com/WebGameLinux/cli/core"
		"github.com/WebGameLinux/cli/resolver"
		"github.com/urfave/cli/v2"
)

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
