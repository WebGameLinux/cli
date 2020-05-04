package core

import (
		"github.com/mitchellh/mapstructure"
		"github.com/urfave/cli/v2"
)

const CommandKey = "Command"

func CreateCommandByMap(data map[string]interface{}) *cli.Command {
		if len(data) == 0 {
				return nil
		}
		cmd := cli.Command{}
		if err := mapstructure.Decode(data, &cmd); err != nil {
				return nil
		}
		return &cmd
}
