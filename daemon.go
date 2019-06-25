package app

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func (a *App) daemon() {
	var nc, stop, debug, restart bool

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
					fmt.Printf("[server: %s, version: %s] is already exists\n\n", a.AppName, a.AppVersion)
					os.Exit(0)
				}
			}
		}
		a.runAsDaemon()
	}

	a.savePidFile()
}

func (a *App) runAsDaemon() {
	if ppid := os.Getppid(); ppid == 1 {
		return
	}

	cmd := exec.Command(GetAbs(os.Args[0]))
	if len(os.Args) > 1 {
		cmd.Args = append(cmd.Args, os.Args[1:]...)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("exec cmd: %+v, err: %v\n\n", cmd, err)
		os.Exit(-1)
	} else {
		fmt.Printf("[server: %s, version: %s] run as daemon success.\n\n", a.AppName, a.AppVersion)
	}

	os.Exit(0)
}

func (a *App) killProcess() {
	pidFile := fmt.Sprintf("%s/run/%s.pid", GetCurrPath(), a.AppName)
	if pidStr, err := ioutil.ReadFile(pidFile); err == nil {
		if pid, _ := strconv.Atoi(string(pidStr)); pid > 0 {
			if p, err := os.FindProcess(pid); err == nil {
				if err = p.Signal(syscall.SIGKILL); err == nil {
					fmt.Printf("[server: %s, version: %s] stoped okay!\n\n", a.AppName, a.AppVersion)
				} else {
					fmt.Printf("kill %s failed, pid: %d, err: %v\n\n", a.AppName, pid, err)
				}
				if err := os.Remove(pidFile); err != nil {
					fmt.Printf("remove pid_file: %s failed. err: %v\n\n", pidFile, err)
				}
			} else {
				fmt.Printf("os find process: %d failed\n\n", pid)
			}
		} else {
			fmt.Printf("pid should greater than 0\n\n")
		}
	} else {
		fmt.Printf("not exists\n\n")
	}
}

func (a *App) savePidFile() {
	pid := []byte(fmt.Sprintf("%d", syscall.Getpid()))
	pidFile := fmt.Sprintf("%s/run/%s.pid", GetCurrPath(), a.AppName)
	if err := ioutil.WriteFile(pidFile, pid, 0644); err != nil {
		fmt.Printf("write pid_file: %s failed, err: %v\n\n", pidFile, err)
	}
}
