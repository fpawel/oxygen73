package app

import (
	"context"
	"github.com/fpawel/oxygen73/internal"
	"github.com/fpawel/oxygen73/internal/cfg"
	"github.com/fpawel/oxygen73/internal/data"
	"github.com/fpawel/oxygen73/internal/gui"
	"github.com/fpawel/oxygen73/internal/pkg/winapi"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"os"
	"os/signal"
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

	// старт ожидания сигнала прерывания ОС
	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-done
		log.Info("приложение закрыто сигналом ОС: прервать все фоновые горутины")
		interrupt()
	}()

	<-ctx.Done()

	log.Debug("прервать все фоновые горутины")
	interrupt()

	log.Debug("остановка горутины, считывающей измерения")
	stopReadMeasurements()

	log.Debug("остановка сервера")
	stopServer()

	log.Debug("закрыть соединение с базой данных")
	log.ErrIfFail(db.Close)

	log.Debug("закрыть окно gui")
	win.SendMessage(winapi.FindWindowClass(internal.WindowClass), win.WM_CLOSE, 0, 0)

	// записать в лог что всё хорошо
	log.Debug("all canceled and closed")
}

var (
	log = structlog.New()
)
