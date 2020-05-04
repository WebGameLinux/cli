package core

import (
		"github.com/mitchellh/mapstructure"
		"github.com/urfave/cli/v2"
		"log"
)

// 命令处理逻辑函数接口
type CommanderHandler interface {
		Handle(*cli.Context) error
}

type BaseCommander struct {
		Command     *cli.Command
		initialized bool
}

type CommanderArguments struct {
		Command *cli.Command
}

func NewCommander(data map[string]interface{}) Commander {
		if len(data) == 0 {
				return &BaseCommander{Command: NewCommand()}
		}
		if v, ok := data[CommandKey]; ok {
				if cmd, ok := v.(*cli.Command); ok {
						return &BaseCommander{
								Command:     cmd,
								initialized: true,
						}
				}
		}
		cmd := CreateCommandByMap(data)
		if cmd == nil {
				return &BaseCommander{Command: nil}
		}
		return &BaseCommander{Command: cmd, initialized: true}
}

func (args *CommanderArguments) Arguments() []cli.Flag {
		return args.Command.Flags
}

func (args *CommanderArguments) SetArguments(lists []cli.Flag) {
		args.Command.Flags = lists
}

func (args *CommanderArguments) AddArgument(arg cli.Flag) {
		args.Command.Flags = append(args.Command.Flags, arg)
}

func NewCommanderArgument(cmd *cli.Command) CommanderArgument {
		return &CommanderArguments{
				Command: cmd,
		}
}

func (commander *BaseCommander) Argument() CommanderArgument {
		if !commander.initialized || commander.Command == nil {
				commander.Init()
		}
		return NewCommanderArgument(commander.Command)
}


// 初始化
func (commander *BaseCommander) Init() {
		if commander.Command != nil && commander.initialized {
				return
		}
		commander.InitCommand()
		commander.initialized = true
}

// 自定义引导
func (commander *BaseCommander) Boot() {
		return
}

// 构造新command
func (commander *BaseCommander) InitCommand() *BaseCommander {
		if commander.Command == nil {
				commander.Command = NewCommand()
		}
		return commander
}

// 必须自行实现注册 覆盖 BaseCommander
func (commander *BaseCommander) Register(app *CmdApp) {
		panic("must impl Register method")
		// Register(commander, app)
}

// 默认commander 自动注册 到 cli 应用的方法
func CommanderRegisterFn(commander Commander, app *CmdApp) {
		if app == nil {
				return
		}
		app.InsertCommand(commander)
}

// 获取 command 对象
func (commander *BaseCommander) GetCommand() *cli.Command {
		if commander.Command == nil {
				commander.InitCommand()
		}
		return commander.Command
}

// 通过 map 设置 command 相关参数
func (commander *BaseCommander) SetProperties(properties map[string]interface{}) {
		if err := mapstructure.Decode(properties, commander.Command); err != nil {
				log.Fatal(err)
		}
}

// 请执行实现 处理逻辑
func (commander *BaseCommander) Handle(ctx *cli.Context) error {
		panic("must impl Handle method")
}

func (commander *BaseCommander) Arguments() (flags []cli.Flag) {
		if len(commander.Command.Flags) == 0 {
				return
		}
		return commander.Command.Flags
}
