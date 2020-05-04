package resolver

type Local string

const (
		ZhLocal Local = "zh"
		EnLocal Local = "en"
)

const CommandEnHelper = `
NAME:
   {{.Name}} {{if ne .Desc "" }}- {{.Desc}}{{end}}

USAGE:
   {{.File}} [global options]
   {{if ne .Usage ""}}
   {{.Usage}}
   {{end}}

GLOBAL OPTIONS:
	{{range $i, $v := .Arguments}}
    --{{$v.Name}} {{$v.Usage}} (default:{{$v.DefaultValue}})
  {{end}}
`

const CommandZhHelper = `
命令名:
   {{.Name}} {{if ne .Desc "" }}- {{.Desc}}{{end}}

用法 :
   {{.File}} [可选参数] 
	 {{if ne .Usage ""}}
   {{.Usage}}
   {{end}}

可选参数:
   {{range $i, $v := .Arguments}}
    --{{$v.Name}}   {{$v.Usage}}   (default:{{$v.DefaultValue}})
  {{end}}
`
