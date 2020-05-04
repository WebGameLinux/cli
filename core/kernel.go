package core

import (
		"github.com/mitchellh/mapstructure"
		"github.com/urfave/cli/v2"
		"log"
		"os"
)

type Runner interface {
		Run() error
}

// 命令自动挂载
type CommandAutoRegister interface {
		Register(app *CmdApp)
}

// 通用属性设置
type MapPropertiesSetter interface {
		SetProperties(map[string]interface{})
}

// 自动初始化内部属性 和 方法
type AutoInit interface {
		Init()
}

// 自主 加载相关信息
type BootAble interface {
		Boot()
}

// 参数命令行
type Commander interface {
		AutoInit
		BootAble
		CommandAutoRegister
		GetCommand() *cli.Command
		Argument() CommanderArgument
}

// 命令行参数器
type CommanderArgument interface {
		Arguments() []cli.Flag
		SetArguments([]cli.Flag)
		AddArgument(cli.Flag)
}

// 命令行应用
type CmdApp struct {
		cli.App
		locked      bool
		initialized bool
		commanders  []Commander
		caches      map[string]int
}

func NewCmdApp() *CmdApp {
		return &CmdApp{
				caches: make(map[string]int),
		}
}

func NewCommand() *cli.Command {
		return &cli.Command{}
}

// 运行
func (cmd *CmdApp) Run() error {
		if !cmd.IsLocked() {
				cmd.Init()
		}
		return cmd.App.Run(os.Args)
}

// 是否已经初始化 , 加载过
func (cmd *CmdApp) IsLocked() bool {
		return cmd.initialized && cmd.locked
}

// 初始化
func (cmd *CmdApp) Init() {
		if cmd.initialized && cmd.locked {
				return
		}
		if len(cmd.commanders) == 0 {
				return
		}
		cmd.LoadCommands(cmd.commanders...).locked = true
		cmd.App.Setup()
}

// 加载
func (cmd *CmdApp) LoadCommands(commanders ...Commander) *CmdApp {
		if len(commanders) == 0 {
				commanders = cmd.commanders
		}
		for i, command := range commanders {
				command.Init()
				command.Boot()
				c := command.GetCommand()
				if cmd.IsLoaded(c.Name) || c.Name == "" {
						continue
				}
				if c.Action == nil {
						if handler, ok := command.(CommanderHandler); ok {
								c.Action = handler.Handle
						}
				}
				cmd.AppendCommand(command, i)
		}
		return cmd
}

func (cmd *CmdApp) Swap(i, j int) {
		name1, name2 := cmd.Commands[i].Name, cmd.Commands[j].Name
		cmd.Commands[i], cmd.Commands[j] = cmd.Commands[j], cmd.Commands[i]
		cmd.caches[name2], cmd.caches[name1] = i, j
}

func (cmd *CmdApp) Less(i, j int) bool {
		var size = cmd.Len()
		if i > size || j > size && i < 0 || j < 0 {
				return false
		}
		name1 := cmd.Commands[i].Name
		name2 := cmd.Commands[j].Name
		return cmd.caches[name1] <= cmd.caches[name2]
}

// 追加命令
func (cmd *CmdApp) AppendCommand(commander Commander, index ...int) {
		if commander == nil || cmd.IsLocked() {
				return
		}
		size := cmd.Len()
		c := commander.GetCommand()
		if len(index) == 0 {
				index = append(index, size)
		}
		if c != nil {
				cmd.Commands = append(cmd.Commands, c)
				i := index[0]
				if i <= size {
						i = size + i
				}
				cmd.caches[c.Name] = i
		}
}

// 命令数量
func (cmd *CmdApp) Len() int {
		return len(cmd.Commands) + 1
}

// 是否已加载
func (cmd *CmdApp) IsLoaded(key string) bool {
		if _, ok := cmd.caches[key]; ok {
				return true
		}
		return false
}

// 添加command
func (cmd *CmdApp) InsertCommand(commander Commander) *CmdApp {
		if commander == nil {
				return cmd
		}
		cmd.commanders = append(cmd.commanders, commander)
		return cmd
}

// 设置 应用属性
func (cmd *CmdApp) SetProperties(properties map[string]interface{}) {
		if err := mapstructure.Decode(properties, cmd); err != nil {
				log.Fatal(err)
		}
}
