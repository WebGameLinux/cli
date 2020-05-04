package resolver

import (
		"fmt"
		"github.com/urfave/cli/v2"
		"regexp"
		"strconv"
		"strings"
)

type StringFlags string

const DefStrFlagDiv = `:`
const DefStrFlagLineDiv = `\n`
const StrFlagHidden rune = '#'
const StrFlagRequired rune = '@'
const DefStrFlagLineRune rune = '\n'
const TypeRegexpPattern = `^\[([a-z]+)\]`

type FlagResolver interface {
		fmt.Stringer
		Row() interface{}
		Args() []cli.Flag
		SetSlice(string) FlagResolver
		GetSlice() string
}

type FlagsStrWrapper struct {
		RowStr StringFlags
		Slice  string
		ArgArr []cli.Flag
}

func Str2FlagTemplate(str string) StringFlags {
		return StringFlags(str)
}

func (s StringFlags) String() string {
		return string(s)
}

func (s StringFlags) Resolver() FlagResolver {
		return NewFlagsStrWrapper(s)
}

func StrTemplateResolver(template string, slice ...string) FlagResolver {
		if len(slice) >= 1 {
				return NewFlagsStrWrapper(template, slice[0])
		}
		return NewFlagsStrWrapper(template)
}

func NewFlagsStrWrapper(str ...interface{}) FlagResolver {
		var argc = len(str)
		if argc == 0 {
				return &FlagsStrWrapper{}
		}
		var resolver *FlagsStrWrapper
		arg := str[0]
		switch arg.(type) {
		case string:
				resolver = &FlagsStrWrapper{RowStr: Str2FlagTemplate(arg.(string))}
		case StringFlags:
				resolver = &FlagsStrWrapper{RowStr: arg.(StringFlags)}
		}
		if argc >= 2 && resolver != nil {
				switch str[1].(type) {
				case string:
						resolver.Slice = str[1].(string)
				}
		}
		return resolver
}

func (wrapper *FlagsStrWrapper) String() string {
		return string(wrapper.RowStr)
}

func (wrapper *FlagsStrWrapper) Row() interface{} {
		return wrapper.RowStr
}

func (wrapper *FlagsStrWrapper) SetSlice(slice string) FlagResolver {
		wrapper.Slice = slice
		return wrapper
}

func (wrapper *FlagsStrWrapper) GetSlice() string {
		if wrapper.Slice == "" {
				return DefStrFlagDiv
		}
		return wrapper.Slice
}

func StrFlagsWrapper(flags StringFlags, slice ...string) *FlagsStrWrapper {
		wrapper := &FlagsStrWrapper{
				RowStr: flags,
				Slice:  DefStrFlagDiv,
		}
		if len(slice) >= 0 {
				wrapper.Slice = slice[0]
		}
		return wrapper
}

func (wrapper *FlagsStrWrapper) Args() (flags []cli.Flag) {
		if wrapper.ArgArr == nil || len(wrapper.ArgArr) == 0 {
				wrapper.Parse()
		}
		return wrapper.ArgArr
}

func (wrapper *FlagsStrWrapper) Parse() {
		// 分解行
		arr := SliceFlags(wrapper.RowStr.String(), DefStrFlagLineDiv)
		if len(arr) == 0 {
				return
		}
		// 解析每一行
		for _, v := range arr {
				argInfoArr := SliceFlags(v, wrapper.GetSlice())
				if len(argInfoArr) <= 0 {
						continue
				}
				arg := BuildFlagByArgInfoStrArr(argInfoArr)
				if arg == nil {
						continue
				}
				wrapper.ArgArr = append(wrapper.ArgArr, arg)
		}
}

func BuildFlagByArgInfoStrArr(info []string) cli.Flag {
		var (
				usage string
				def   string
				flag  cli.Flag
				argc  = len(info)
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
				required := Str2Bool(data["required"])
				hidden := Str2Bool(data["hidden"])
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
						flag = &cli.StringFlag{
								Name:     values["name"].(string),
								Aliases:  values["alias"].([]string),
								Usage:    usage,
								Value:    def,
								Required: required,
								Hidden:   hidden,
						}
				case "bool":
						fallthrough
				case "b":
						fallthrough
				case "boolean":
						flag = &cli.BoolFlag{
								Name:     values["name"].(string),
								Aliases:  values["alias"].([]string),
								Usage:    usage,
								Value:    Str2Bool(def),
								Required: required,
								Hidden:   hidden,
						}
				case "number":
						fallthrough
				case "int":
						fallthrough
				case "integer":
						flag = &cli.IntFlag{
								Name:     values["name"].(string),
								Aliases:  values["alias"].([]string),
								Usage:    usage,
								Value:    Str2Int(def),
								Required: required,
								Hidden:   hidden,
						}
				case "integers":
				case "intArr":
						flag = &cli.IntSliceFlag{
								Name:     values["name"].(string),
								Aliases:  values["alias"].([]string),
								Usage:    usage,
								Value:    Str2IntSlice(def),
								Required: required,
								Hidden:   hidden,
						}
				case "float":
						fallthrough
				case "double":
						flag = &cli.Float64Flag{
								Name:     values["name"].(string),
								Aliases:  values["alias"].([]string),
								Usage:    usage,
								Value:    Str2Float64(def),
								Required: required,
								Hidden:   hidden,
						}
				}
		}
		return flag
}

func Str2IntSlice(str string) (arr *cli.IntSlice) {
		if !strings.Contains(str, ",") {
				return cli.NewIntSlice(Str2Int(str))
		}
		var (
				nums   []int
				strArr = FilterArrString(strings.Split(str, ","))
		)
		for _, v := range strArr {
				n := Str2Int(v)
				nums = append(nums, n)
		}
		if len(nums) > 0 {
				return cli.NewIntSlice(nums...)
		}
		return
}

func Str2Float64(str string) float64 {
		if num, err := strconv.ParseFloat(str, 64); err == nil {
				return num
		}
		return 0
}

func Str2Bool(str string) bool {
		if str == "1" || str == "true" ||
				str == "True" || str == "TRUE" ||
				str == "ok" || str == "yes" ||
				str == "OK" || str == "YES" {
				return true
		}
		return false
}

func Str2Int(str string) int {
		if num, err := strconv.Atoi(str); err == nil {
				return num
		}
		return 0
}

func Str2IntArray(nums string) []int {
		var arr []int
		var v []string
		if nums == "" || nums == " " {
				return arr
		}
		if strings.Contains(nums, ",") {
				v = strings.SplitN(nums, ",", -1)
				v = FilterArrString(v)
		}
		if len(v) == 0 && strings.Contains(nums, " ") {
				v = strings.SplitN(nums, " ", -1)
				v = FilterArrString(v)
		}
		if len(v) < 0 {
				return arr
		}
		for _, n := range v {
				n = strings.Trim(n, " ")
				in := Str2Int(n)
				if in == 0 && (n != "0" && n != "0.0") {
						continue
				}
				arr = append(arr, in)
		}
		return arr
}

func GetFlagType(ty string) (map[string]string, bool) {
		if ty == "" || ty == " " {
				return nil, false
		}
		data := make(map[string]string)
		data["hidden"] = "false"
		data["required"] = "false"
		ty = strings.Trim(ty, " ")
		reg := regexp.MustCompile(TypeRegexpPattern)
		runes := []rune(ty)
		if runes[0] == rune(StrFlagRequired) {
				data["required"] = "true"
				ty = strings.Replace(ty, string(StrFlagRequired), "", 1)
		}
		if runes[0] == rune(StrFlagHidden) {
				data["hidden"] = "true"
				ty = strings.Replace(ty, string(StrFlagHidden), "", 1)
		}
		subs := reg.FindAllStringSubmatch(ty, -1)
		if len(subs) < 1 {
				data["type"] = "string"
				data["flags"] = ty
				return data, strings.Contains(ty, "-")
		}
		data["type"] = strings.Trim(subs[0][1], " ")
		data["flags"] = strings.Replace(ty, subs[0][0], "", -1)
		return data, true
}

func GetFlagNames(names string) (map[string]interface{}, bool) {
		if names == "" || names == " " {
				return nil, false
		}
		var alias []string
		names = strings.Trim(names, " ")
		data := make(map[string]interface{})

		if !strings.Contains(names, ",") {
				names = strings.Replace(names, "-", "", -1)
				data["name"] = names
		} else {
				ns := FilterArrString(strings.Split(names, ","))
				size := len(ns)
				if size <= 0 {
						return nil, false
				}
				for _, v := range ns {
						v = strings.Trim(v, " ")
						if strings.Contains(v, "--") && len(v) >= 4 {
								data["name"] = strings.Replace(v, "--", "", 1)
								continue
						}
						if strings.Contains(v, "-") && len(v) >= 2 {
								v = strings.Replace(v, "-", "", 1)
								if len(v) > 2 {
										continue
								}
								alias = append(alias, v)
						}
				}
		}
		data["alias"] = alias
		return data, true
}

func SliceFlags(row string, slice string) (arr []string) {
		if row == "" {
				return
		}
		if slice == "" {
				slice = DefStrFlagLineDiv
		}
		row = strings.Trim(row, " ")
		if strings.Contains(row, slice) {
				return FilterArrString(strings.Split(row, slice))
		}
		if slice == DefStrFlagLineDiv && strings.Contains(row, string(DefStrFlagLineRune)) {
				arr = strings.SplitN(row, string(DefStrFlagLineRune), -1)
		}
		if len(arr) < 0 {
				return arr
		}
		return FilterArrString(arr)
}

func FilterArrString(arr []string) []string {
		if len(arr) == 0 {
				return arr
		}
		for i, v := range arr {
				if v == "" || v == " " {
						arr = DeleteByIndex(arr, i)
				}
				if v == "\\n" || v == "\\t" || v == "\\r" {
						arr = DeleteByIndex(arr, i)
				}
				if v == "\\t\\r" || v == "\\r\\n" || v == "\\t\\n" {
						arr = DeleteByIndex(arr, i)
				}
		}
		return arr
}

func DeleteByIndex(arr []string, index int) []string {
		var size = len(arr)
		var caps = cap(arr)
		var rArr []string
		if index > size && caps < index {
				return rArr
		}
		if index == 0 {
				return arr[index+1:]
		}
		if index+1 >= size {
				return arr[:index-1]
		}
		rArr = arr[index+1:]
		if len(rArr) == 0 {
				return arr[:index]
		}
		arr = append(arr[:index], rArr...)
		return arr
}
