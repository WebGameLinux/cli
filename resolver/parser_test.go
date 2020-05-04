package resolver

import (
		"testing"
)

const InitConfFileName = "../examples/daemon/.daemon.ini"

func TestIniParser_GetName(t *testing.T) {
		app := NewIniParser()
		if err := app.ParserIniFile(InitConfFileName); err != nil {
				t.Error(err)
		}
		// fmt.Printf("%+v", app.IM)
		if app.GetName("web", "pidFile") == "" {
				t.Error("解析获取值失败")
		}
}
