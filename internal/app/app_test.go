package app

import (
	"context"
	"github.com/fpawel/oxygen73/internal"
	"github.com/fpawel/oxygen73/internal/cfg"
	"github.com/fpawel/oxygen73/internal/data"
	"github.com/fpawel/oxygen73/internal/gui"
	"github.com/fpawel/oxygen73/internal/pkg"
	"github.com/fpawel/oxygen73/internal/pkg/winapi"
	"github.com/jmoiron/sqlx"
	"github.com/lxn/win"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestAppMain(t *testing.T) {
	pkg.InitLog()

	// общий контекст приложения с прерыванием
	ctx, interrupt := context.WithCancel(context.Background())

	// открыть конфиг
	cfg.Open(log)

	// соединение с базой данных
	db := data.OpenDev()

	// сервер
	server := newServer(mainSvcHandler{db: db})

	// старт сервера
	go log.ErrIfFail(server.Serve, "detail", "`failed to run main service`")

	// старт цикла оконных сообщений окна связи с gui
	go func() {
		runtime.LockOSThread()
		gui.Handle()
		interrupt()
	}()

	// старт горутины, считывающей измерения
	stopReadMeasurements := testRunReadMeasurements(ctx, db)

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
	win.SendMessage(winapi.FindWindowClass(internal.WindowClass), win.WM_CLOSE, 0, 0)

	// записать в лог что всё хорошо
	log.Debug("all canceled and closed")
}

func testRunReadMeasurements(ctx context.Context, db *sqlx.DB) context.CancelFunc {

	data.MustLastParty(db)
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		var ms data.Measurements
		for {
			if ctx.Err() != nil {
				return
			}
			m := data.Measurement{
				Temperature: randV(),
				Pressure:    randV(),
				Humidity:    randV(),
				StoredAt:    time.Now(),
			}

			for i := 0; i < 50; i++ {
				m.Places[i] = randV()
			}
			gui.Measurement(m)

			ms = append(ms, m)

			c := cfg.Get()
			if len(ms) >= c.Public.SaveMeasurementsCount {
				log.ErrIfFail(func() error {
					return data.SaveMeasurements(ms, db)
				})
				ms = nil
			}
			pause(ctx.Done(), time.Second)
		}
	}()

	return func() {
		wg.Wait()
	}
}

func randV() float64 {
	return math.Floor(rand.Float64()*100) / 100
}
