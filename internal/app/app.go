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
	"io"
	"os"
	"os/signal"
	"syscall"
)

func Main() {

	if winapi.IsWindow(winapi.FindWindow(internal.WindowClass)) {
		log.Fatalln("window class", internal.WindowClass, "already exists")
	}

	// общий контекст приложения с прерыванием
	ctx, interrupt := context.WithCancel(context.Background())

	// открыть конфиг
	cfg.Open(log)

	// соединение с базой данных
	db := data.OpenProd()

	// сервер
	server := newServer(mainSvcHandler{
		db: db,
	})

	// старт сервера
	go log.ErrIfFail(server.Serve, "detail", "`failed to run main service`")

	// старт цикла оконных сообщений окна связи с gui
	go func() {
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
	log.ErrIfFail(server.Stop, "detail", "`failed to stop main service`")

	log.Debug("закрыть соединение с базой данных")
	log.ErrIfFail(db.Close)

	log.Debug("закрыть окно gui")
	win.SendMessage(gui.HWndSource(), win.WM_CLOSE, 0, 0)

	// записать в лог что всё хорошо
	log.Debug("all canceled and closed")
}

func GUIWriter() io.Writer {
	return guiWriter{}
}

type guiWriter struct{}

func (x guiWriter) Write(p []byte) (int, error) {
	go gui.W{}.WriteConsole(string(p))
	return len(p), nil
}

var (
	log = structlog.New()
)
