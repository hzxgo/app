package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hzxgo/cors"
	"github.com/hzxgo/log"
	"github.com/hzxgo/mysql"
)

// 通用App
type App struct {
	AppName    string         // app name
	AppVersion string         // app version
	AppPort    string         // app listen port
	Env        string         // app run env
	SessionOn  bool           // session switch
	Cors       []string       // cors
	Tasks      []SubGoroutine // sub goroutine
}

// interface
// 子携程需实现该接口
type SubGoroutine interface {
	GetTaskName() string
	GoroutineStart() error
	GoroutineStop() error
}

// ---------------------------------------------------------------------------------------------------------------------

func NewApp(appName, appVersion, appPort, env string) *App {
	return &App{
		AppName:    appName,
		AppVersion: appVersion,
		AppPort:    appPort,
		Env:        env,
		Tasks:      make([]SubGoroutine, 0, 1),
	}
}

func (a *App) Init() *gin.Engine {
	a.daemon()

	localIp, _ := GetLocalAddr()

	fmt.Println()
	log.Infof("------------------------------------------")
	log.Infof("app_name: %s", a.AppName)
	log.Infof("app_version: %s", a.AppVersion)
	log.Infof("local_address: %s", localIp)
	log.Infof("start-up_time: %s", time.Now().Format("2006-01-02 15:04:05"))
	log.Info("------------------------------------------")

	r := gin.Default()

	// 设置跨域
	if len(a.Cors) > 0 {
		r.Use(cors.Default(a.Cors))
	}

	a.SafeExit()

	return r
}

// 安全退出
func (a *App) SafeExit() {
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(
			signalChan, os.Kill, os.Interrupt,
			syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		signalMsg := <-signalChan
		log.Warnf("receive signal %v, app is closing", signalMsg)

		a.FreeResource()
		os.Exit(1)
	}()
}

// 添加子携程
func (a *App) AddSubGoroutine(tasks ...SubGoroutine) {
	if len(tasks) > 0 {
		a.Tasks = append(a.Tasks, tasks...)
	}
}

// 启动所有子携程
func (a *App) StartAllSubGoroutine() error {
	for _, v := range a.Tasks {
		if err := v.GoroutineStart(); err != nil {
			return err
		}
		log.Infof("task: %v start success", v.GetTaskName())
	}
	return nil
}

// 关闭所有子携程
func (a *App) StopAllSubGoroutine() {
	for _, v := range a.Tasks {
		if err := v.GoroutineStop(); err != nil {
			log.Errorf("task: %v stop err: %v", v.GetTaskName(), err)
		} else {
			log.Infof("task: %v stop success", v.GetTaskName())
		}
	}
}

// 释放资源
func (a *App) FreeResource() {
	a.StopAllSubGoroutine()
	mysql.FreeDB()
}
