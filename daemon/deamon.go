package daemon

import (
		"context"
		"flag"
		"fmt"
		"github.com/WebGameLinux/cli/resolver"
		"io/ioutil"
		"os"
		"os/exec"
		"path"
		"path/filepath"
		"runtime"
		"strconv"
		"strings"
		"time"
)

const (
		FlagName           = "d"
		ConfKeyDaemonOn    = "deamonOn"
		ConfKeyPidFile     = "pidFile"
		ConfKeyLogFile     = "logFile"
		ConfKeyGroupOn     = "groupOn"
		InitConfFileName   = ".daemon.ini"
		FlagUsage          = "run app as a daemon with -d=true : (default:false)"
		FlagScopeName      = "name"
		FlagScopeNameUsage = "run app daemon with name : --name=appName (default:global)"
		DefScopeName       = "app"
		FlagStopName       = "stop"
		FlagStopNameUsage  = "stop app daemon  : --stop (default:false)"
		ArgDaemonOn        = "-d=true"
		ArgDaemonOff       = "-d=false"
		EnvDaemonPidFile   = "Daemon.PidFile"
)

// 启动
type InitDaemonConf struct {
		DaemonOn bool
		PidFile  string
		LogFile  string
		GroupOn  int
		App      string
}

// 后台启动
func init() {
		var (
				on   *bool
				app  *string
				stop *bool
				conf *InitDaemonConf
		)
		if v := flag.Lookup(FlagName); v != nil {
				panic("please don't set flag: '" + FlagName + "'  ")
		} else {
				on = flag.Bool(FlagName, false, FlagUsage)
		}
		app = flag.String(FlagScopeName, DefScopeName, FlagScopeNameUsage)
		stop = flag.Bool(FlagStopName, false, FlagStopNameUsage)
		if !flag.Parsed() {
				flag.Parse()
		}
		exits, file := GetInitConfFile()
		if exits && file != "" {
				conf = parseIni(file, *app)
		}
		// 暂停
		if *stop {
				pidFile := ""
				if conf != nil && conf.PidFile != "" {
						pidFile = conf.PidFile
				}
				stopProcess(pidFile, *app)
				os.Exit(0)
		}
		// 启动 (命令行指定,或者配置文件指定)
		if *on || (conf != nil && conf.DaemonOn) {
				start(true, conf)
		}
}

// 后台启动
func start(on bool, conf *InitDaemonConf) {
		if !on {
				return
		}
		name := "app"
		exists := false
		args := os.Args[1:]
		for i := 0; i < len(args); i++ {
				if args[i] == ArgDaemonOff {
						return
				}
				if args[i] == ArgDaemonOn {
						args[i] = ArgDaemonOff
						exists = true
						break
				}
		}
		if !exists {
				args = append(args, ArgDaemonOff)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cmd := exec.CommandContext(ctx, os.Args[0], args...)
		if err := cmd.Start(); err != nil {
				fmt.Println(err)
				cancel()
				return
		}
		if conf != nil && conf.App != "" {
				name = conf.App
		}
		fmt.Printf("%s running [PID] : %d\n", name, cmd.Process.Pid)
		startAfter(cmd.Process.Pid, conf)
		os.Exit(0)
}

// 解析ini
func parseIni(file string, scope string) *InitDaemonConf {
		var (
				conf   = InitDaemonConf{}
				parser = resolver.NewIniParser()
		)
		if err := parser.ParserIniFile(file); err != nil {
				return &conf
		}
		v := parser.GetName(scope, ConfKeyDaemonOn)
		if v != "" {
				conf.DaemonOn = resolver.Str2Bool(v)
		}
		v = parser.GetName(scope, ConfKeyPidFile)
		if v != "" {
				conf.PidFile = v
		}
		v = parser.GetName(scope, ConfKeyLogFile)
		if v != "" {
				conf.LogFile = v
		}
		v = parser.GetName(scope, ConfKeyGroupOn)
		if v != "" {
				conf.GroupOn = resolver.Str2Int(v)
		}
		conf.App = scope
		return &conf
}

// 获取配置文件
func GetInitConfFile() (bool, string) {
		var (
				dir     string
				err     error
				absPath string
				info    os.FileInfo
				file    = os.Args[0]
		)
		if absPath, err = filepath.Abs(file); err != nil {
				fmt.Println(err.Error())
				os.Exit(-1)
		}
		dir = filepath.Dir(absPath)
		file = path.Join(dir, InitConfFileName)
		if info, err = os.Stat(file); err != nil {
				if os.IsNotExist(err) {
						return false, ""
				}
				fmt.Println(err)
				return false, ""
		}
		if info.IsDir() || info.Size() == 0 {
				return false, ""
		}
		return true, file
}

// 保存pid
func savePid(pid int, num int, filename ...string) {
		var (
				err    error
				v      os.FileInfo
				size   int64
				fs     *os.File
				f      int
				pidStr string
				pidArr []string
		)
		if len(filename) <= 0 {
				return
		}
		file := filename[0]
		if v, err = os.Stat(file); err != nil {
				if !os.IsNotExist(err) {
						fmt.Println(err)
						return
				}
				v = nil
		}
		if v != nil {
				size = v.Size()
		}
		if fs, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_SYNC|os.O_APPEND, os.ModePerm); err != nil {
				fmt.Println(err)
				return
		}
		defer closeFile(fs)
		// 通过file 中记录的 pid 杀死 进程
		if size > 0 {
				var (
						buf = make([]byte, size)
				)
				if n, err := fs.Read(buf); err == nil && n > 0 {
						pidStr = strings.Trim(string(buf), " ")
				}
				if pidStr != "" && strings.Contains(pidStr, ",") {
						pidArr = resolver.FilterArrString(strings.SplitN(pidStr, ",", -1))
				}
		}
		// 关闭多余进程
		if num <= 0 && size > 0 {
				if len(pidArr) > 0 {
						for _, p := range pidArr {
								if pid, err := strconv.Atoi(p); err == nil && pid != 1 && pid != 0 {
										kill(pid, true)
								}
						}
				} else {
						if pid, err := strconv.Atoi(pidStr); err == nil && pid != 1 && pid != 0 {
								kill(pid, true)
						}
				}
		}
		// 判断保存模式
		pidStr = fmt.Sprintf("%d", pid)
		if num <= 0 {
				f = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		} else {
				f = os.O_WRONLY | os.O_CREATE | os.O_APPEND
				if size > 0 {
						pidStr = "," + pidStr
				}
		}
		// 启动进程数超过限制
		if len(pidArr) > num {
				kill(pid, true)
				return
		}
		// 	打开文件
		if fs, err = os.OpenFile(file, f, os.ModePerm); err != nil {
				return
		}
		defer closeFile(fs)
		// 记录 pid
		_, err = fs.WriteString(pidStr)
}

// 关闭文件
func closeFile(fs *os.File) {
		_ = fs.Close()
}

// 启动之后
func startAfter(pid int, config *InitDaemonConf) {
		if config != nil {
				if config.PidFile != "" {
						savePid(pid, config.GroupOn, config.PidFile)
				}
				return
		}
		file := os.Getenv(EnvDaemonPidFile)
		if file != "" {
				savePid(pid, 0, file)
		}
}

// 停止进程
func stopProcess(pidFile string, app string) {
		var (
				pid  int
				file string
		)
		if pidFile != "" && fileExits(pidFile, 0) {
				file = pidFile
				if buf, err := ioutil.ReadFile(pidFile); err == nil {
						pid = resolver.Str2Int(string(buf))
				}
				// 移除 pidfile
				defer delFile(file)
		} else {
				appPs := "'" + os.Args[0] + ".*--name=" + app + "'"
				switch runtime.GOOS {
				case "darwin":
						fallthrough
				case "linux":
						pid = psUnix(appPs)
				case "windows":
						pid = psWin(app)
				}
		}
		if pid == 0 || pid == 1 {
				fmt.Println("pid not exits")
				return
		}
		if !kill(pid, true) {
				return
		}
		fmt.Printf("app : %s , stoped \n", app)
}

// 通用 进程killer
func kill(pid int, force ...bool) bool {
		if pid == 0 || pid == 1 {
				return false
		}
		// 是否强杀
		isForceKill := false
		if len(force) > 0 && force[0] {
				isForceKill = true
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		pidStr := fmt.Sprintf("%d", pid)
		switch runtime.GOOS {
		case "darwin":
				fallthrough
		case "linux":
				signalArg := "-15"
				if isForceKill {
						signalArg = "-9"
				}
				if err := exec.CommandContext(ctx, "kill", signalArg, pidStr).Run(); err != nil {
						fmt.Println(err)
				}
				return true
		case "windows":
				signalArg := " "
				if isForceKill {
						signalArg = "-f"
				}
				if err := exec.CommandContext(ctx, "taskkill", "/pid", pidStr, "-t", signalArg).Run(); err != nil {
						fmt.Println(err)
				}
				return true
		}
		return false
}

// 查看 进程 pid
func psWin(app string) int {
		// @todo test
		cmd := exec.Command("tasklist", "|findstr ", app)
		if buf, err := cmd.Output(); err == nil {
				res := resolver.FilterArrString(strings.Split(string(buf), " "))
				if len(res) < 2 {
						return 0
				}
				return resolver.Str2Int(res[1])
		}
		return 0
}

// 删除文件
func delFile(file string) {
		_ = os.Remove(file)
}

// Linux 系统 查找 进程
func psUnix(app string) int {
		cmd := exec.Command("ps", "-ef | grep -E ", app, `grep -v "grep" | awk '{print $2}'`)
		if buf, err := cmd.Output(); err == nil {
				return resolver.Str2Int(string(buf))
		}
		return 0
}

// 文件是否在
func fileExits(file string, minSize int) bool {
		var (
				v   os.FileInfo
				err error
		)
		if v, err = os.Stat(file); err != nil {
				if os.IsNotExist(err) {
						return false
				}
				if os.IsPermission(err) {
						return false
				}
		}
		if v.IsDir() {
				return false
		}
		if v.Size() < int64(minSize) {
				return false
		}
		return true
}
