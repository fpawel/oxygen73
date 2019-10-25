package app

import (
	"context"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/oxygen73/internal/cfg"
	"github.com/fpawel/oxygen73/internal/data"
	"github.com/fpawel/oxygen73/internal/gui"
	"github.com/jmoiron/sqlx"
	"sync"
	"time"
)

func runReadMeasurements(ctx context.Context, db *sqlx.DB) context.CancelFunc {

	var wg sync.WaitGroup
	wg.Add(1)

	data.MustLastParty(db)

	comPort := comport.NewPort(func() comport.Config {
		return comport.Config{
			Baud:        115200,
			ReadTimeout: time.Millisecond,
			Name:        cfg.Get().Main.Comport,
		}
	})

	go func() {
		defer wg.Done()
		var measurements data.Measurements
		comportName := cfg.Get().Main.Comport
	workerLoop:
		for {

			if ctx.Err() != nil {
				log.Info("close worker because of context: " + ctx.Err().Error())
				return
			}
			c := cfg.Get()
			if c.Main.Comport != comportName {
				log.ErrIfFail(comPort.Close)
				comportName = c.Main.Comport
			}

			reader := comPort.NewResponseReader(ctx, c.Main.Comm)
			var measurement data.Measurement
			for n := 0; n < 5; n++ {
				valuesCount := 10
				if n == 0 {
					valuesCount = 12
				}
				// получить значения напряжений 50 каналов, температуры и давления
				values, err := modbus.Read3BCDs(log, reader, 101+modbus.Addr(n), 0, valuesCount)

				if ctx.Err() != nil {
					log.Info("close worker because of context: " + ctx.Err().Error())
					return
				}

				if err != nil {
					gui.StatusErr(err)
					pause(ctx.Done(), c.Main.Comm.ReadTimeout())
					continue workerLoop
				}

				gui.StatusOk("связь установлена")

				if n == 0 {
					measurement.Temperature = values[10]
					measurement.Pressure = values[11]
				}
				copy(measurement.Places[n*10:(n+1)*10], values[:10])
			}
			measurement.StoredAt = time.Now()
			measurements = append(measurements, measurement)
			if len(measurements) >= c.SaveMeasurementsCount {
				saveMeasurements := measurements
				measurements = nil
				wg.Add(1)
				go func() {
					if err := data.SaveMeasurements(saveMeasurements, db); err != nil {
						log.PrintErr("не удалось сохранить измерения", "reason", err)
					}
					gui.NewMeasurements(saveMeasurements)
					wg.Done()
				}()
			}
		}
	}()

	return func() {
		wg.Wait()
		log.ErrIfFail(comPort.Close)
	}
}
