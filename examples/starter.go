package examples

import "github.com/WebGameLinux/cli/resolver"

const (
		FlagsExpress = `
[bool]--help : show help menus : false
--conf : config file path  : ./config.ini
--name : set process name  :  cli
[bool]--deamon : process run in background : true
`
)

func GetCommander() *resolver.SimpleCommanderApp {
		resolver.SimpleCommandLine.BootHandler = func(app *resolver.SimpleCommanderApp) {
				app.AddFlagsByStrExpress(FlagsExpress)
		}
		return resolver.SimpleCommandLine
}


