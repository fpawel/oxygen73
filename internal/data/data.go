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

type ProductVoltage struct {
	StoredAt        time.Time
	SeriesCreatedAt time.Time
	Place           int
	SerialNumber    int
	Tension         float64
}

type Ambient struct {
	StoredAt        time.Time
	SeriesCreatedAt time.Time
	Temperature     float64
	Pressure        float64
	Humidity        float64
}

type timeSql struct {
	Year        int        `db:"year"`
	Month       time.Month `db:"month"`
	Day         int        `db:"day"`
	Hour        int        `db:"hour"`
	Minute      int        `db:"minute"`
	Second      int        `db:"second"`
	Millisecond int        `db:"millisecond"`
}

func (x timeSql) Time() time.Time {
	return time.Date(
		x.Year, x.Month, x.Day,
		x.Hour, x.Minute, x.Second,
		x.Millisecond*int(time.Millisecond/time.Nanosecond),
		time.Local)
}

func ProductVoltageUpdatedAt() time.Time {
	var t timeSql
	err := db.Get(&t, `SELECT * FROM product_voltage_updated_at `)
	if err == nil || err == sql.ErrNoRows {
		return t.Time()
	}
	panic(err)
}

func AddProductVoltage(place, serial int, tension float64) {
	mu.Lock()
	defer mu.Unlock()
	x := ProductVoltage{
		StoredAt:     time.Now(),
		Place:        place,
		SerialNumber: serial,
		Tension:      tension,
	}
	if len(productVoltageSeries) == 0 {
		//var t timeSql
	}

	productVoltageSeries = append(productVoltageSeries, x)
}

func AddAmbient(x Ambient) {
	mu.Lock()
	defer mu.Unlock()
	ambientSeries = append(ambientSeries, x)
}

func SaveAndCleanCache() {
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

func queryInsertProductVoltages() string {
	queryStr := `INSERT INTO product_voltage(place, serial_number, tension, stored_at) VALUES `
	for i, a := range productVoltageSeries {

		s := fmt.Sprintf("(%d, %d, %v,", a.Place, a.SerialNumber, a.Tension) +
			"julianday(STRFTIME('%Y-%m-%d %H:%M:%f','" +
			a.StoredAt.Format("2006-01-02 15:04:05.000") + "')))"
		if i < len(productVoltageSeries)-1 {
			s += ", "
		}
		queryStr += s
	}
	return queryStr
}

func queryInsertAmbient() string {
	queryStr := `INSERT INTO ambient(temperature, pressure, humidity, stored_at) VALUES `
	for i, a := range ambientSeries {

		s := fmt.Sprintf("(%v, %v, %v,", a.Temperature, a.Pressure, a.Humidity) +
			"julianday(STRFTIME('%Y-%m-%d %H:%M:%f','" +
			a.StoredAt.Format("2006-01-02 15:04:05.000") + "')))"
		if i < len(productVoltageSeries)-1 {
			s += ", "
		}
		queryStr += s
	}
	return queryStr
}

var (
	db = func() *sqlx.DB {
		db := gohelp.OpenSqliteDBx(filepath.Join(internal.DataDir(), "series.sqlite"))
		db.MustExec(SQLCreate)
		return db
	}()
	productVoltageSeries []ProductVoltage
	ambientSeries        []Ambient
	mu                   sync.Mutex
)

func init() {
	go func() {
		t := time.NewTicker(time.Minute)
		for {
			<-t.C
			SaveAndCleanCache()
		}
	}()
}
