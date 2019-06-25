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

type App struct {
	AppName    string         // app name
	AppVersion string         // app version
	SessionOn  bool           // session switch
	Cors       []string       // cors
	SubGo      []SubGoroutine // sub goroutine
}

type SubGoroutine interface {
	GoroutineStart() error
	GoroutineStop() error
}

// ---------------------------------------------------------------------------------------------------------------------

func NewApp(appName, appVersion string) *App {
	return &App{
		AppName:    appName,
		AppVersion: appVersion,
		SubGo:      make([]SubGoroutine, 0, 1),
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
	if len(a.Cors) > 0 {
		r.Use(cors.Default(a.Cors))
	}

	a.SafeExit()

	return r
}

func (a *App) SafeExit() {
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan,
			os.Kill, os.Interrupt,
			syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		signalMsg := <-signalChan
		log.Warnf("receive signal %v, app is closing", signalMsg)

		a.StopAllSubGoroutine()
		a.FreeResource()
		os.Exit(1)
	}()
}

func (a *App) AppendSubGoroutine(subGo ...SubGoroutine) {
	if len(subGo) > 0 {
		a.SubGo = append(a.SubGo, subGo...)
	}
}

func (a *App) StartAllSubGoroutine() error {
	for _, v := range a.SubGo {
		if err := v.GoroutineStart(); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) StopAllSubGoroutine() {
	for _, v := range a.SubGo {
		_ = v.GoroutineStop()
	}
}

func (a *App) FreeResource() {
	mysql.FreeDB()
}
