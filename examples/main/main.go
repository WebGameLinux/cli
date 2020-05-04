package main

import (
		"fmt"
		"github.com/WebGameLinux/cli/core"
		"github.com/WebGameLinux/cli/examples"
		"github.com/urfave/cli/v2"
		"log"
		"os"
)

// urfave/cli 方式
func test() {
		var language string
		language = "测试"
		app := &cli.App{
				Flags: []cli.Flag{
						&cli.StringFlag{
								Name:        "lang",
								Value:       "english",
								Usage:       "language for the greeting",
								Destination: &language,
						},
				},
				Action: func(c *cli.Context) error {
						name := "someone"
						if c.NArg() > 0 {
								name = c.Args().Get(0)
						}
						if language == "spanish" {
								fmt.Println("Hola", name)
						} else {
								fmt.Println("Hello", name)
						}
						return nil
				},
		}
		app.Commands = []*cli.Command{}
		err := app.Run(os.Args)

		if err != nil {
				log.Fatal(err)
		}
}

func Version() {
		app := core.NewCmdApp()
		cmd := examples.NewVersionCmd()
		cmd.Register(app)
		app.Run()
}

func main() {
		app := examples.GetCommander()
		app.Run()
		if app.GetBool("help") {
				app.PrintUsage()
		} else {
				fmt.Println(app.GetBool("deamon"))
				fmt.Println(app.Get("conf"))
				fmt.Println(app.Get("name"))
		}

		//cmd := exec.CommandContext(context.TODO(), "ping","www.baidu.com")
		//err := cmd.Start()
		//fmt.Println(err)
		//fmt.Println("[PID]", cmd.Process.Pid)
		//os.Exit(0)
}
