package resolver

import (
		"fmt"
		"strings"
		"testing"
)

const (
		ZhArgs = `
@[string]--config,-c: 配置文件路径 : file.txt
[bool]--on :  是否开启 : false 
--version,-V,-ver : 显示版本号 : 0.0.1
`
		EnArgs = `
@[string]--config,-c: test : file.txt
[bool]--on :  open on button : false 
--version,-V,-ver : show version : 0.0.1
`
)

var tests = []string{
		"[]",
		"@[string]--config,-c: user : default",
		"#[123]",
		"[bool]--bool,-b",
		"[number]-n,--number",
		"[int]--int,-i",
		"[json]--json,-j",
		"[file]--file,-f",
		"[path]--path,-p:user,default",
}

func TestGetFlagType(t *testing.T) {

		var str string
		str = strings.Join(tests, "\\n")
		if len(StrTemplateResolver(str).Args()) <= 1 {
				t.Error("解析带换行命令参数异常")
		}
		args := StrTemplateResolver(ZhArgs).Args()
		fmt.Println(args)
		if len(args) != 3 {
				t.Error("解析中文命令参数异常")
		}
		args = StrTemplateResolver(EnArgs).Args()
		if len(args) != 3 {
				t.Error("解析英文命令参数异常")
		}
		fmt.Println(args)
}

func GetCommander(express string) *SimpleCommanderApp {
		SimpleCommandLine.BootHandler = func(app *SimpleCommanderApp) {
				app.AddFlagsByStrExpress(express)
		}
		return SimpleCommandLine
}

func TestGetHelp(t *testing.T) {
		app:=GetCommander(ZhArgs)
		app.Name = "test"
		app.Desc = "test commander"
		app.PrintUsage()
}
