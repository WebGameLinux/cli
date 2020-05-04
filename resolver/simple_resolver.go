package resolver

import (
		"bytes"
		"flag"
		"fmt"
		"github.com/WebGameLinux/cli/core"
		"github.com/mitchellh/mapstructure"
		"io"
		"log"
		"os"
		"path"
		"strings"
		"text/template"
)

type SimpleResolver interface {
		Args() []*flag.Flag
		core.AutoInit
		core.BootAble
		GetUsage() func()
		GetName() string
		HelpBoot()
		Parse() bool
}

type SimpleCommanderAppApi interface {
		Get(string) string
		GetInt(string) int
		GetNum(string) int
		GetBool(string) bool
		GetBigInt(string) int64
		GetFloat(string) float64
		GetArgN(int) string
		GetOsArgs() []string
		GetOsArgN(int) string
		Run()
}

type SimpleCommanderApp struct {
		Usage       interface{}                  // 用户帮助说明
		AutoParse   bool                         // 是否初始时自动解析
		Parsed      bool                         // 是否已经解析
		Name        string                       // 命令应用名
		Desc        string                       // 命令应用描述
		Lang        Local                        // 语言
		ArgumentMap map[string]*CommanderFlagDto // 参数存储容器
		BootHandler func(app *SimpleCommanderApp)
}

// 命令参数 数据结构
type CommanderFlagDto struct {
		Name         string
		Usage        string
		DefaultValue interface{}
		Value        interface{}
		RowType      string
}

// 命令行应用 体 数据结构
type CommanderAppDto struct {
		Arguments []*CommanderFlagDto
		Usage     string
		Name      string
		Desc      string
		Version   string
		File      string
}

func NewCommanderFlagDto() *CommanderFlagDto {
		return &CommanderFlagDto{}
}

// 初始化
func (flagDto *CommanderFlagDto) InitFlagDto(name, rowType, usage string, value interface{}, defValue interface{}) *CommanderFlagDto {
		flagDto.RowType = rowType
		flagDto.Name = name
		flagDto.Usage = usage
		flagDto.Value = value
		flagDto.DefaultValue = defValue
		return flagDto
}

func NewCommanderAppDto() *CommanderAppDto {
		return &CommanderAppDto{File: path.Base(os.Args[0])}
}

func (appDto *CommanderAppDto) InitDto(app *SimpleCommanderApp) *CommanderAppDto {
		for _, flagDto := range app.ArgumentMap {
				appDto.Arguments = append(appDto.Arguments, flagDto)
		}
		appDto.Name = app.Name
		if v, ok := app.Usage.(string); ok {
				appDto.Usage = v
		}
		appDto.Desc = app.Desc
		if appDto.Version == "" {
				appDto.Version = "0.0.1"
		}
		if appDto.File == "" {
				appDto.File = path.Base(os.Args[0])
		}
		if app.Name == "" || !strings.Contains(appDto.File,appDto.Name) {
				appDto.Name = appDto.File
		}
		return appDto
}

func (app *SimpleCommanderApp) GetValue(key string) interface{} {
		if !flag.Parsed() {
				app.Parse()
		}
		if v, ok := app.ArgumentMap[key]; ok {
				return v.Value
		}
		return nil
}

func (app *SimpleCommanderApp) GetOsArgs() []string {
		return os.Args
}

func (app *SimpleCommanderApp) GetOsArgN(i int) string {
		return GetOsArgN(i)
}

func (app *SimpleCommanderApp) Get(key string) string {
		val := app.GetValue(key)
		if v, ok := val.(*string); ok {
				return *v
		}
		if v, ok := val.(string); ok {
				return v
		}
		return ""
}

func (app *SimpleCommanderApp) GetArgN(i int) string {
		return flag.CommandLine.Arg(i)
}

func (app *SimpleCommanderApp) GetInt(key string) int {
		val := app.GetValue(key)
		if v, ok := val.(*int); ok {
				return *v
		}
		if v, ok := val.(int); ok {
				return v
		}
		return 0
}

func (app *SimpleCommanderApp) GetBigInt(key string) int64 {
		val := app.GetValue(key)
		if v, ok := val.(*int64); ok {
				return *v
		}
		if v, ok := val.(int64); ok {
				return v
		}
		return 0
}

func (app *SimpleCommanderApp) GetNum(key string) int {
		return app.GetInt(key)
}

func (app *SimpleCommanderApp) GetBool(key string) bool {
		val := app.GetValue(key)

		if v, ok := val.(*bool); ok {
				return *v
		}
		if v, ok := val.(bool); ok {
				return v
		}
		return false
}

func (app *SimpleCommanderApp) GetFloat(key string) float64 {
		val := app.GetValue(key)
		if v, ok := val.(*float64); ok {
				return *v
		}
		if v, ok := val.(float64); ok {
				return v
		}
		return 0
}

func (app *SimpleCommanderApp) Args() (args []*flag.Flag) {
		if app.argc() == 0 {
				app.Init()
		}
		for key, _ := range app.ArgumentMap {
				args = append(args, flag.Lookup(key))
		}
		return args
}

func (app *SimpleCommanderApp) GetName() string {
		return app.Name
}

func (app *SimpleCommanderApp) Boot() {
		if app.BootHandler == nil {
				panic("must self impl Boot method")
		}
		app.BootHandler(app)
}

func (app *SimpleCommanderApp) Init() {
		if app.ArgumentMap == nil {
				app.ArgumentMap = make(map[string]*CommanderFlagDto)
		}
		app.Boot()
		if app.Lang == "" {
				app.Lang = EnLocal
		}
		if !app.HasArg("help") {
				app.HelpBoot()
		}
		flag.Usage = app.GetUsage()
		if app.AutoParse {
				app.Parse()
		}
}

func (app *SimpleCommanderApp) HasArg(key string) bool {
		if _, ok := app.ArgumentMap[key]; ok {
				return true
		}
		return false
}

func (app *SimpleCommanderApp) Parse() bool {
		if !flag.Parsed() {
				flag.Parse()
		}
		if app.argc() == 0 && !app.AutoParse {
				app.Init()
		}
		app.Parsed = true
		return app.Parsed
}

func (app *SimpleCommanderApp) GetUsage() func() {
		if app.Usage == nil {
				return app.help()
		}
		switch app.Usage.(type) {
		case func():
				return app.Usage.(func())
		case string:
				return app.HelpWrapper(app.Usage.(string))
		case func() string:
				return app.HelpWrapper(app.Usage.(func() string)())
		}
		return app.help()
}

func (app *SimpleCommanderApp) PrintUsage() {
		app.GetUsage()()
}

func (app *SimpleCommanderApp) Usage2Str() string {
		var (
				err     error
				fs      *os.File
				tmpFile = ".tmp"
				stdout  = os.Stdout
				buf     []byte
				size    int64 = 1024
		)
		if fs, err = os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0777); err != nil {
				return ""
		}
		defer ReleaseStdout(stdout, fs)
		os.Stdout = fs
		app.PrintUsage()
		_, err = fs.Seek(1, io.SeekStart)
		// 计算buffer 大小
		if stat, err := fs.Stat(); err == nil {
				size = stat.Size()
		}
		bufSize := int64(len(buf))
		// buffer 大小调整
		if size > bufSize || size+1 < bufSize {
				buf = make([]byte, size+1)
		}
		if n, err := fs.Read(buf); err != nil || 0 == n {
				return ""
		}
		return string(buf)
}

func (app *SimpleCommanderApp) HelpWrapper(usage string) func() {
		return func() {
				fmt.Println(usage)
		}
}

func (app *SimpleCommanderApp) HelpBoot() {
		var usage string
		if app.Lang == ZhLocal {
				usage = "展示命令菜单"
		} else {
				usage = "show help"
		}
		app.AddFlag("bool", "help", usage, false)
}

func (app *SimpleCommanderApp) AddFlag(ty string, name string, usage string, defs ...interface{}) {
		if app.HasArg(name) {
				return
		}
		if key, v := FlagAppend(ty, name, usage, defs...); v != nil {
				if dto, ok := v.(*CommanderFlagDto); ok {
						app.ArgumentMap[key] = dto
				}
		}
}

func (app *SimpleCommanderApp) argc() int {
		return len(app.ArgumentMap)
}

func (app *SimpleCommanderApp) Run() {
		app.Parse()
}

func (app *SimpleCommanderApp) help() func() {
		if app.argc() == 0 {
				app.Init()
		}
		var usage string
		if app.Lang == ZhLocal {
				usage = CommandZhHelper
		} else {
				usage = CommandEnHelper
		}
		usage = TemplateParse(app.Name, usage, NewCommanderAppDto().InitDto(app))
		return func() {
				fmt.Print(usage)
		}
}

func (app *SimpleCommanderApp) AddFlagsByStrExpress(tpl string) {
		arr := SliceFlags(tpl, DefStrFlagLineDiv)
		if len(arr) == 0 {
				return
		}
		// 解析每一行
		for _, v := range arr {
				argInfoArr := SliceFlags(v, DefStrFlagDiv)
				if len(argInfoArr) <= 0 {
						continue
				}
				dto := BuildFlagDtoByArgInfoStrArr(argInfoArr)
				if dto == nil {
						continue
				}
				app.AddFlag(dto.RowType, dto.Name, dto.Usage, dto.DefaultValue)
		}
}

func ParseFlagsExpress(express string) []*CommanderFlagDto {
		var lists []*CommanderFlagDto
		arr := SliceFlags(express, DefStrFlagLineDiv)
		if len(arr) == 0 {
				return nil
		}
		// 解析每一行
		for _, v := range arr {
				argInfoArr := SliceFlags(v, DefStrFlagDiv)
				if len(argInfoArr) <= 0 {
						continue
				}
				dto := BuildFlagDtoByArgInfoStrArr(argInfoArr)
				if dto == nil {
						continue
				}
				lists = append(lists, dto)
		}
		return lists
}

func BuildFlagDtoByArgInfoStrArr(info []string) *CommanderFlagDto {
		var (
				usage string
				def   string
				argc  = len(info)
				dto   = NewCommanderFlagDto()
		)
		if argc < 1 {
				return nil
		}
		if argc >= 2 {
				usage = info[1]
		}
		if argc >= 3 {
				def = strings.Trim(info[2], " ")
		}
		if data, ok := GetFlagType(info[0]); ok {
				ty := data["type"]
				flags := data["flags"]
				if flags == "" {
						return nil
				}
				values, ok := GetFlagNames(flags)
				if !ok {
						return nil
				}
				switch ty {
				case "string":
						fallthrough
				case "path":
						fallthrough
				case "file":
						fallthrough
				case "str":
						fallthrough
				case "json":
						fallthrough
				case "map":
						fallthrough
				case "any":
						dto.InitFlagDto(values["name"].(string), ty, usage, nil, def)
				case "bool":
						fallthrough
				case "b":
						fallthrough
				case "boolean":
						dto.InitFlagDto(values["name"].(string), ty, usage, nil, Str2Bool(def))
				case "number":
						fallthrough
				case "int":
						fallthrough
				case "integer":
						dto.InitFlagDto(values["name"].(string), ty, usage, nil, Str2Int(def))
				case "bigInt":
						dto.InitFlagDto(values["name"].(string), ty, usage, nil, int64(Str2Int(def)))
				case "float":
						fallthrough
				case "double":
						dto.InitFlagDto(values["name"].(string), ty, usage, nil, Str2Float64(def))
				}
		}
		return dto
}

func NewSimpleCommanderApp() SimpleResolver {
		return &SimpleCommanderApp{}
}

func NewSimpleCommanderAppByMap(data map[string]interface{}) *SimpleCommanderApp {
		app := &SimpleCommanderApp{}
		if err := mapstructure.Decode(data, app); err != nil {
				log.Fatal(err)
		}
		if app.ArgumentMap == nil {
				app.ArgumentMap = make(map[string]*CommanderFlagDto)
		}
		return app
}

func FlagAppend(typeName string, name string, usage string, defs ...interface{}) (string, interface{}) {
		argc := len(defs)
		dto := NewCommanderFlagDto()
		switch typeName {
		case "bool":
				fallthrough
		case "boolean":
				fallthrough
		case "yes":
				fallthrough
		case "no":
				var (
						value        bool
						defaultValue bool
				)
				if argc > 0 {
						if v, ok := defs[0].(bool); ok {
								defaultValue = v
						}
				}
				flag.BoolVar(&value, name, defaultValue, usage)
				return name, dto.InitFlagDto(name, typeName, usage, &value, defaultValue)
		case "string":
				fallthrough
		case "str":
				fallthrough
		case "file":
				fallthrough
		case "path":
				var (
						value        string
						defaultValue string
				)
				if argc > 0 {
						if v, ok := defs[0].(string); ok {
								defaultValue = v
						}
				}
				flag.StringVar(&value, name, defaultValue, usage)
				return name, dto.InitFlagDto(name, typeName, usage, &value, defaultValue)
		case "int":
				fallthrough
		case "integer":
				fallthrough
		case "number":
				var (
						value        int
						defaultValue int
				)
				if argc > 0 {
						if v, ok := defs[0].(int); ok {
								defaultValue = v
						}
				}
				flag.IntVar(&value, name, defaultValue, usage)
				return name, dto.InitFlagDto(name, typeName, usage, &value, defaultValue)
		case "bigInt":
				var (
						value        int64
						defaultValue int64
				)
				flag.Int64Var(&value, name, defaultValue, usage)
				return name, dto.InitFlagDto(name, typeName, usage, &value, defaultValue)
		case "float":
				var (
						value        float64
						defaultValue float64
				)
				if argc > 0 {
						if v, ok := defs[0].(float64); ok {
								defaultValue = v
						}
				}
				flag.Float64Var(&value, name, defaultValue, usage)

				return name, dto.InitFlagDto(name, typeName, usage, &value, defaultValue)
		default:
				var (
						value        string
						defaultValue string
				)
				if argc > 0 {
						if v, ok := defs[0].(string); ok {
								defaultValue = v
						}
				}
				flag.StringVar(&value, name, defaultValue, usage)
				return name, dto.InitFlagDto(name, typeName, usage, &value, defaultValue)
		}
}

func TemplateParse(tplName string, tplContent string, v interface{}, funcMaps ...template.FuncMap) string {
		var funcMap template.FuncMap
		if v == nil {
				return ""
		}
		if len(funcMaps) > 0 {
				funcMap = funcMaps[0]
		}
		tpl, err := template.New(tplName).Funcs(funcMap).Parse(tplContent)
		if err != nil {
				log.Print(err)
				return ""
		}
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, v); err != nil {
				log.Print(err)
				return ""
		}
		return buf.String()
}

func ReleaseStdout(stdout *os.File, tmp *os.File) {
		os.Stdout = stdout
		file := tmp.Name()
		_ = tmp.Close()
		_ = os.Remove(file)
}

var SimpleCommandLine = &SimpleCommanderApp{}

func Run() {
		SimpleCommandLine.Run()
}

func BindBootHandler(handler func(app *SimpleCommanderApp)) {
		SimpleCommandLine.BootHandler = handler
}

func GetValue(key string) interface{} {
		SimpleCommandLine.Parse()
		return SimpleCommandLine.GetValue(key)
}

func GetOsArgs() []string {
		return os.Args
}

func GetOsArgN(i int) string {
		var argc = len(os.Args)
		if argc <= i || i < 0 {
				return ""
		}
		return os.Args[i]
}

func Get(key string) string {
		val := SimpleCommandLine.GetValue(key)
		if v, ok := val.(*string); ok {
				return *v
		}
		if v, ok := val.(string); ok {
				return v
		}
		return ""
}

func GetArgN(i int) string {
		return flag.CommandLine.Arg(i)
}

func GetInt(key string) int {
		val := SimpleCommandLine.GetValue(key)
		if v, ok := val.(*int); ok {
				return *v
		}
		if v, ok := val.(int); ok {
				return v
		}
		return 0
}

func GetBigInt(key string) int64 {
		val := SimpleCommandLine.GetValue(key)
		if v, ok := val.(*int64); ok {
				return *v
		}
		if v, ok := val.(int64); ok {
				return v
		}
		return 0
}

func GetNum(key string) int {
		return SimpleCommandLine.GetInt(key)
}

func GetBool(key string) bool {
		val := SimpleCommandLine.GetValue(key)

		if v, ok := val.(*bool); ok {
				return *v
		}
		if v, ok := val.(bool); ok {
				return v
		}
		return false
}

func GetFloat(key string) float64 {
		val := SimpleCommandLine.GetValue(key)
		if v, ok := val.(*float64); ok {
				return *v
		}
		if v, ok := val.(float64); ok {
				return v
		}
		return 0
}

func Args() (args []*flag.Flag) {
		if SimpleCommandLine.argc() == 0 {
				SimpleCommandLine.Init()
		}
		for key, _ := range SimpleCommandLine.ArgumentMap {
				args = append(args, flag.Lookup(key))
		}
		return args
}

func GetName() string {
		return SimpleCommandLine.Name
}
