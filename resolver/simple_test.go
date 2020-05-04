package resolver

import (
		"testing"
)

const (
		FlagsExpress = `
[bool]--help : show help menus : false
--conf : config file path  : ./config.ini
--name : set process name  :  app
[bool]--deamon : process run in background : true
`
)

func ExampleNewSimpleCommanderApp() {
		SimpleCommandLine.BootHandler = func(app *SimpleCommanderApp) {
				app.AddFlagsByStrExpress(FlagsExpress)
		}
		// SimpleCommandLine.Usage2Str()
		//SimpleCommandLine.PrintUsage()
		SimpleCommandLine.Parse()
		SimpleCommandLine.Get("name")
}

func TestParseFlagsExpress(t *testing.T) {
		var total = 4
		var arr = ParseFlagsExpress(FlagsExpress)
		for _, v := range arr {
				if v.Name == "" {
						t.Error("解析字符串表达式name异常")
				}
				if v.Usage == "" {
						t.Error("解析字符串表达式usage异常")
				}
				if v.DefaultValue == nil {
						t.Error("解析字符串表达式default异常")
				}
		}
		if len(arr) < total {
				t.Error("解析字符串表达失败")
		}
		ExampleNewSimpleCommanderApp()
}

func BenchmarkParseFlagsExpress(b *testing.B) {
		for i := 0; i < b.N; i++ {
				arr := ParseFlagsExpress(FlagsExpress)
				for _, v := range arr {
						if v.Name == "" {
								b.Error("解析字符串表达式name异常")
						}
						if v.Usage == "" {
								b.Error("解析字符串表达式usage异常")
						}
						if v.DefaultValue == nil {
								b.Error("解析字符串表达式default异常")
						}
				}
		}
}

func BenchmarkSimpleCommanderApp_Usage2Str(b *testing.B) {
		SimpleCommandLine.BootHandler = func(app *SimpleCommanderApp) {
				app.AddFlagsByStrExpress(FlagsExpress)
		}
		var (
				size    int
				curSize int
		)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
				str := SimpleCommandLine.Usage2Str()
				curSize = len(str)
				if size == 0 {
						size = curSize
				}
				if curSize == 0 {
						b.Error("输出字符串usage失败")
				}
				if size != curSize {
						b.Error("输出字符串usage内容不全")
				}
		}
		//SimpleCommandLine.PrintUsage()
		//SimpleCommandLine.Parse()
		//SimpleCommandLine.Get("name")
}
