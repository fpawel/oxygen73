package app

import (
	"context"
	"encoding/binary"
	"fmt"
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

	comPort := new(comport.Port)
	comPortHum := new(comport.Port)

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

			if err := comPort.SetConfig(comport.Config{
				Baud:        115200,
				ReadTimeout: time.Millisecond,
				Name:        cfg.Get().Main.Comport,
			}); err != nil {
				err = merry.Append(err, comportName).Append("стенд")
				gui.StatusComportErr(err)
				pause(ctx.Done(), time.Second)
				continue
			}

			if err := comPortHum.SetConfig(comport.Config{
				Baud:        9600,
				ReadTimeout: time.Millisecond,
				Name:        cfg.Get().Hum.Comport,
			}); err != nil {
				err = merry.Append(err, comportHumName).Append("датчик влажности")
				gui.StatusComportHumErr(err)
				pause(ctx.Done(), time.Second)
				continue
			}

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
				comportHumName = c.Hum.Comport
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
				var hum, temp int16

				_, err := modbus.Read3(log, readerHum, 16, 0x0102, 2,
					func(_, response []byte) (string, error) {
						temp = int16(binary.BigEndian.Uint16(response[3:5]))
						hum = int16(binary.BigEndian.Uint16(response[5:7]))
						return fmt.Sprintf("T=%d,H=%d", temp, hum), nil
					})
				if err != nil {
					err = merry.Append(err, comportHumName).Append("датчик влажности")
					gui.StatusComportHumErr(err)
					pause(ctx.Done(), time.Second)
					return
				}
				gui.StatusComportHumOk(comportHumName + ": датчик влажности: связь установлена")
				measurement.Humidity = float64(hum) / 100.
				measurement.Temperature = float64(temp) / 100.
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
				pause(ctx.Done(), time.Second)
				continue workerLoop
			}
			wgHum.Wait()
			measurement.StoredAt = time.Now()
			measurements = append(measurements, measurement)
			if len(measurements) >= c.SaveMeasurementsCount {
				saveMeasurements(measurements, db, ctx)
				measurements = nil
			}
		}
	}()

	return func() {
		wg.Wait()
		log.ErrIfFail(comPort.Close)
	}
}

func saveMeasurements(measurements data.Measurements, db *sqlx.DB, ctx context.Context) {
	if err := data.SaveMeasurements(measurements, db); err != nil {
		log.PrintErr("не удалось сохранить измерения", "reason", err)
		return
	}
	var bucketID int64
	if err := db.GetContext(ctx, &bucketID, `SELECT bucket_id FROM last_bucket`); err != nil {
		log.PrintErr(merry.Append(err, "can't get last bucket id"))
		return
	}
	gui.NewMeasurements(bucketID, measurements)
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
		//me.Temperature = values[10]
		me.Pressure = values[11]
	}
	copy(me.Places[n*10:(n+1)*10], values[:10])
	return nil
}
