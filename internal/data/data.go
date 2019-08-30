package data

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/gohelp"
	"github.com/fpawel/oxygen73/internal"
	"github.com/jmoiron/sqlx"
	"path/filepath"
	"sync"
	"time"
)

//go:generate go run github.com/fpawel/gohelp/cmd/sqlstr/...

func AddProductVoltage(place, serial int, tension float64) {
	mu.Lock()
	defer mu.Unlock()
	productVoltageSeries = append(productVoltageSeries, productVoltageSample{
		StoredAt:        time.Now(),
		Place:           place,
		SerialNumber:    serial,
		Tension:         tension,
		SeriesCreatedAt: getCurrentSeriesCreatedAt(),
	})
}

func AddAmbient(temperature, pressure, humidity float64) {
	mu.Lock()
	defer mu.Unlock()
	ambientSeries = append(ambientSeries, ambientSample{
		StoredAt:        time.Now(),
		SeriesCreatedAt: getCurrentSeriesCreatedAt(),
		Temperature:     temperature,
		Pressure:        pressure,
		Humidity:        humidity,
	})
}

func Save() {
	mu.Lock()
	defer mu.Unlock()

	if len(productVoltageSeries) > 0 {
		db.MustExec(queryInsertProductVoltages())
		productVoltageSeries = nil
	}

	if len(ambientSeries) > 0 {
		db.MustExec(queryInsertAmbient())
		ambientSeries = nil
	}
}

type productVoltageSample struct {
	StoredAt        time.Time
	SeriesCreatedAt time.Time
	Place           int
	SerialNumber    int
	Tension         float64
}

type ambientSample struct {
	StoredAt        time.Time
	SeriesCreatedAt time.Time
	Temperature     float64
	Pressure        float64
	Humidity        float64
}

func lastSavedProductVoltage() (productVoltageSample, bool) {
	var x struct {
		StoredAt        string  `db:"stored_at_str"`
		SeriesCreatedAt string  `db:"series_created_at_str"`
		Place           int     `db:"place"`
		SerialNumber    int     `db:"serial_number"`
		Tension         float64 `db:"tension"`
	}
	err := db.Get(&x, `SELECT stored_at_str, series_created_at_str, serial_number, place, tension FROM product_voltage_updated_at`)
	switch err {
	case nil:
		return productVoltageSample{
			StoredAt:        parseTime(x.StoredAt),
			SeriesCreatedAt: parseTime(x.SeriesCreatedAt),
			Place:           x.Place,
			SerialNumber:    x.SerialNumber,
			Tension:         x.Tension,
		}, true
	case sql.ErrNoRows:
		return productVoltageSample{}, false
	default:
		panic(err)
	}
}

func getCurrentSeriesCreatedAt() time.Time {
	if len(productVoltageSeries) > 0 {
		return productVoltageSeries[len(productVoltageSeries)-1].SeriesCreatedAt
	}
	if y, f := lastSavedProductVoltage(); f {
		d := time.Since(y.StoredAt)
		if d < 5*time.Minute {
			return y.SeriesCreatedAt
		}
	}
	return time.Now()
}

func queryInsertProductVoltages() string {
	queryStr := `INSERT INTO product_voltage(place, serial_number, tension, stored_at, series_created_at) VALUES `
	for i, a := range productVoltageSeries {

		s := fmt.Sprintf("(%d, %d, %v,", a.Place, a.SerialNumber, a.Tension) +
			formatTimeAsQuery(a.StoredAt) + "," +
			formatTimeAsQuery(a.SeriesCreatedAt) + ")"
		if i < len(productVoltageSeries)-1 {
			s += ", "
		}
		queryStr += s
	}
	return queryStr
}

func queryInsertAmbient() string {
	queryStr := `INSERT INTO ambient(temperature, pressure, humidity, stored_at, series_created_at) VALUES `
	for i, a := range ambientSeries {

		s := fmt.Sprintf("(%v, %v, %v,", a.Temperature, a.Pressure, a.Humidity) +
			formatTimeAsQuery(a.StoredAt) + "," +
			formatTimeAsQuery(a.SeriesCreatedAt) + ")"
		if i < len(productVoltageSeries)-1 {
			s += ", "
		}
		queryStr += s
	}
	return queryStr
}

func parseTime(sqlStr string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04:05.000", sqlStr, time.Now().Location())
	if err != nil {
		panic(err)
	}
	return t
}
func formatTimeAsQuery(t time.Time) string {
	return "julianday(STRFTIME('%Y-%m-%d %H:%M:%f','" +
		t.Format(parseTimeFormat) + "'))"
}

const parseTimeFormat = "2006-01-02 15:04:05.000"

var (
	db = func() *sqlx.DB {
		db := gohelp.OpenSqliteDBx(filepath.Join(internal.DataDir(), "series.sqlite"))
		db.MustExec(SQLCreate)
		return db
	}()
	productVoltageSeries []productVoltageSample
	ambientSeries        []ambientSample
	mu                   sync.Mutex
)

func init() {
	go func() {
		t := time.NewTicker(time.Minute)
		for {
			<-t.C
			Save()
		}
	}()
}
