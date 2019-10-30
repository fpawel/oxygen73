package app

import (
	"context"
	"github.com/fpawel/oxygen73/internal"
	"github.com/fpawel/oxygen73/internal/cfg"
	"github.com/fpawel/oxygen73/internal/data"
	"github.com/fpawel/oxygen73/internal/gui"
	"github.com/fpawel/oxygen73/internal/pkg/winapi"
	"github.com/powerman/structlog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
)

func Main() {

	if winapi.IsWindow(winapi.FindWindowClass(internal.WindowClass)) {
		log.Fatalln("window class", internal.WindowClass, "already exists")
	}

	// общий контекст приложения с прерыванием
	ctx, interrupt := context.WithCancel(context.Background())

	// открыть конфиг
	cfg.Open(log)

	// соединение с базой данных
	db := data.OpenProd()

	// старт сервера
	stopServer := runServer(db)

	// старт цикла оконных сообщений окна связи с gui
	go func() {
		runtime.LockOSThread()
		gui.Handle()
		interrupt()
	}()

	// старт горутины, считывающей измерения
	stopReadMeasurements := runReadMeasurements(ctx, db)

	if len(os.Getenv("OXYGEN73_DEV_MODE")) != 0 {
		log.Debug("waiting system signal because of OXYGEN73_DEV_MODE=" + os.Getenv("OXYGEN73_DEV_MODE"))
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		sig := <-done
		log.Debug("system signal: " + sig.String())
	} else {
		cmd := exec.Command(filepath.Join(filepath.Dir(os.Args[0]), "oxygen73gui.exe"))
		log.ErrIfFail(cmd.Start)
		log.ErrIfFail(cmd.Wait)
		log.Debug("gui was closed.")
	}

	log.Debug("прервать все фоновые горутины")
	interrupt()

	log.Debug("остановка горутины, считывающей измерения")
	stopReadMeasurements()

	log.Debug("остановка сервера")
	stopServer()

	log.Debug("закрыть соединение с базой данных")
	log.ErrIfFail(db.Close)

	// записать в лог что всё хорошо
	log.Debug("all canceled and closed")
}

var (
	log = structlog.New()
)
