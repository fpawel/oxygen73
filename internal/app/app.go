package app

import (
	"context"
	"github.com/fpawel/oxygen73/internal/cfg"
	"github.com/fpawel/oxygen73/internal/data"
	"github.com/fpawel/oxygen73/internal/guiclient"
	"github.com/powerman/structlog"
	"os"
	"os/signal"
	"syscall"
)

func Main() {

	// общий контекст приложения с прерыванием
	ctx, interrupt := context.WithCancel(context.Background())

	// старт ожидания сигнала прерывания ОС
	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-done
		log.Info("interrupt: signal close accepted")
		interrupt()
	}()

	// открыть конфиг
	cfg.Open(log)

	// соединение с базой данных
	db := data.OpenProd()

	// сервер
	server := newServer(mainSvcHandler{
		db: db,
		interruptApp: func() {
			log.Info("interrupt: handler")
			interrupt()
		},
	})

	// старт сервера
	go log.ErrIfFail(server.Serve, "detail", "`failed to run main service`")

	// старт горутины, считывающей измерения
	stopReadMeasurements := runReadMeasurements(ctx, db)

	// ожидание прерывания общего контекста приложения ctx
	// контекст может быть прерван:
	//	- в обработчике соединений сервера
	//	- в обработчике сигнала прерывания ОС
	<-ctx.Done()

	// прервать все фоновые горутины
	interrupt()

	// остановка горутины, считывающей измерения
	stopReadMeasurements()

	// закрыть клиент gui
	log.ErrIfFail(guiclient.Close, "detail", "`failed to close gui client`")

	// остановка сервера
	log.ErrIfFail(server.Stop, "detail", "`failed to stop main service`")

	// закрыть соединение с базой данных
	log.ErrIfFail(db.Close)

	// записать в лог что всё хорошо
	log.Debug("all canceled and closed")
}

var (
	log = structlog.New()
)
