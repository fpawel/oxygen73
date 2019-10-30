package app

import (
	"context"
	"encoding/binary"
	"github.com/ansel1/merry"
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

	if _, err := data.LastParty(ctx, db); err != nil {
		panic(err)
	}

	comPort := comport.NewPort(func() comport.Config {
		return comport.Config{
			Baud:        115200,
			ReadTimeout: time.Millisecond,
			Name:        cfg.Get().Main.Comport,
		}
	})

	comPortHum := comport.NewPort(func() comport.Config {
		return comport.Config{
			Baud:        9600,
			ReadTimeout: time.Millisecond,
			Name:        cfg.Get().Hum.Comport,
		}
	})

	go func() {
		defer wg.Done()
		var (
			measurements                data.Measurements
			comportName, comportHumName string
		)
		{
			c := cfg.Get()
			comportName = c.Main.Comport
			comportHumName = c.Hum.Comport
		}

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
			if c.Hum.Comport != comportHumName {
				log.ErrIfFail(comPortHum.Close)
				comportName = c.Hum.Comport
			}

			reader := comPort.NewResponseReader(ctx, c.Main.Comm())
			readerHum := comPortHum.NewResponseReader(ctx, c.Hum.Comm())
			var (
				measurement data.Measurement
				wgHum       sync.WaitGroup
			)
			wgHum.Add(1)
			go func() {
				defer wgHum.Done()
				v, err := modbus.Read3UInt16(log, readerHum, 16, 0x0103, binary.BigEndian)
				if err != nil {
					err = merry.Append(err, comportHumName).Append("датчик влажности")
					gui.StatusComportHumErr(err)
					return
				}
				gui.StatusComportHumOk(comportHumName + ": датчик влажности: связь установлена")
				measurement.Humidity = float64(v) / 100.
			}()

			for n := 0; n < 5; n++ {
				err := readBlock(n, reader, &measurement)
				if err == nil {
					gui.StatusComportOk(comportName + ": стенд: связь установлена")
					continue
				}
				if ctx.Err() != nil {
					log.Info("close worker because of context: " + ctx.Err().Error())
					return
				}
				err = merry.Append(err, comportName).Append("стенд")
				gui.StatusComportErr(err)
				pause(ctx.Done(), c.Main.Comm().ReadTimeout())
				continue workerLoop
			}

			wgHum.Wait()
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

func readBlock(n int, reader modbus.ResponseReader, me *data.Measurement) error {
	valuesCount := 10
	if n == 0 {
		valuesCount = 12
	}
	// получить значения напряжений 50 каналов, температуры и давления
	values, err := modbus.Read3BCDs(log, reader, 101+modbus.Addr(n), 0, valuesCount)
	if err != nil {
		return err
	}
	if n == 0 {
		me.Temperature = values[10]
		me.Pressure = values[11]
	}
	copy(me.Places[n*10:(n+1)*10], values[:10])
	return nil
}
