package app

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// 服务后台运行
func (a *App) daemon() {
	var nc, stop, debug, restart bool

	// 启动参数解析
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Printf("These are the options you can pass:\n")
		fmt.Printf(" -help     -- show this message\n")
		fmt.Printf(" -version  -- print server version and exit\n")
		fmt.Printf(" -debug    -- run to forward and print log to console\n")
		fmt.Printf(" -nc       -- run to background and write log to file\n")
		fmt.Printf(" -stop     -- stop the server\n")
		fmt.Printf(" -restart  -- restart the server\n")
	}

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	// check args
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-debug":
			debug = true
		case "-nc":
			nc = true
		case "-help":
			flag.Usage()
			os.Exit(0)
		case "-version":
			fmt.Println(a.AppVersion)
			os.Exit(0)
		case "-stop":
			stop = true
		case "-restart":
			restart = true
		default:
		}
	}

	if !debug && !stop && !nc && !restart {
		flag.Usage()
		os.Exit(0)
	}

	if debug && !nc {
		return
	}

	root := GetCurrPath()
	if !IsDirExist(root + "/run") {
		if err := os.Mkdir(root+"/run", 0755); err != nil {
			fmt.Printf("mkdir %s/run | %v\n\n", root, err)
			os.Exit(0)
		}
	}

	if stop {
		a.killProcess()
		os.Exit(0)
	}

	if nc {
		pidFile := fmt.Sprintf("%s/run/%s.pid", root, a.AppName)
		if pidStr, err := ioutil.ReadFile(pidFile); err == nil {
			if pid, _ := strconv.Atoi(string(pidStr)); pid > 0 {
				if _, err := os.FindProcess(pid); err == nil {
					fmt.Printf("[server: %s, env: %v, version: %s] is already exists\n", a.AppName, a.Env, a.AppVersion)
					os.Exit(0)
				}
			}
		}
		a.runAsDaemon()
	}

	if restart {
		a.killProcess()
		a.runAsDaemon()
	}

	a.savePidFile()
}

// 服务以守护进程方式运行
func (a *App) runAsDaemon() {
	if ppid := os.Getppid(); ppid == 1 {
		return
	}

	cmd := exec.Command(GetAbs(os.Args[0]), "-nc")
	if err := cmd.Start(); err != nil {
		fmt.Printf("exec cmd: %+v, err: %v\n\n", cmd, err)
		os.Exit(-1)
	} else {
		fmt.Printf("[server: %s, env: %v, version: %s] run as daemon success.\n\n", a.AppName, a.Env, a.AppVersion)
	}

	os.Exit(0)
}

// 杀死进程（向进程发Kill信号）
func (a *App) killProcess() {
	var pid int

	// git pid
	pidFile := fmt.Sprintf("%s/run/%s.pid", GetCurrPath(), a.AppName)
	pidString, err := ioutil.ReadFile(pidFile)
	if pid, _ = strconv.Atoi(string(pidString)); pid <= 0 {
		if pid, _ = a.getPid(a.AppPort, a.AppName); pid <= 0 {
			fmt.Printf("kill %s failed, err: %v\n\n", a.AppName, err)
			return
		}
	}

	if p, err := os.FindProcess(pid); err == nil {
		if err = p.Signal(syscall.SIGKILL); err == nil {
			fmt.Printf("[server: %s, env: %v, version: %s] stoped okay!\n\n", a.AppName, a.Env, a.AppVersion)
		} else {
			fmt.Printf("kill %s failed, pid: %d, err: %v\n\n", a.AppName, pid, err)
		}
		if err := os.Remove(pidFile); err != nil {
			fmt.Printf("%v\n\n", err)
		}
	} else {
		fmt.Printf("os find process: %d failed\n\n", pid)
	}
}

// 获取进程ID
func (a *App) getPid(port string, serverName string) (int, error) {
	cmd := fmt.Sprintf(`netstat -tnlp | grep %s `, port)
	output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return 0, err
	}

	outputString := strings.TrimSpace(string(output))
	outputSlice := strings.Split(outputString, "\n")

	for _, v := range outputSlice {
		// line_format: [tcp 0 0 127.0.0.1:5432 0.0.0.0:* LISTEN 19436/postgres]
		line := strings.Fields(v)
		if length := len(line); length > 0 {
			if portNameSlice := strings.Split(line[length-1], "/"); len(portNameSlice) == 2 {
				findPid, _ := strconv.Atoi(strings.TrimSpace(portNameSlice[0]))
				findName := strings.TrimSpace(portNameSlice[1])
				if findName == serverName {
					return findPid, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("pid not found")
}

// 保存PID号
func (a *App) savePidFile() {
	pid := []byte(fmt.Sprintf("%d", syscall.Getpid()))
	pidFile := fmt.Sprintf("%s/run/%s.pid", GetCurrPath(), a.AppName)
	if err := ioutil.WriteFile(pidFile, pid, 0644); err != nil {
		fmt.Printf("write pid_file: %s failed, err: %v\n\n", pidFile, err)
	}
}
