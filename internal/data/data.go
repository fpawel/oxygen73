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
	Time    time.Time
	Place   int
	Number  int
	Tension float64
}

type Ambient struct {
	Time        time.Time
	Temperature float64
	Pressure    float64
	Humidity    float64
}

func UpdatedAt() (time.Time, bool){
	var t time.Time
	err := db.Get(&t, `SELECT time FROM product_voltage ORDER BY time DESC LIMIT 1`)
	switch err {
	case nil:
		return t, true
	case sql.ErrNoRows:
		return t,false
	default:
		panic(err)
	}
}

func AddProductVoltages( xs []ProductVoltage){
	mu.Lock()
	defer mu.Unlock()
	productVoltageSeries = append(productVoltageSeries, xs...)
}

func AddAmbient( x Ambient){
	mu.Lock()
	defer mu.Unlock()
	ambientSeries = append(ambientSeries, x)
}

func SaveAndCleanCache(){
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
	queryStr := `INSERT INTO product_voltage(place, number, tension, time) VALUES `
	for i, a := range productVoltageSeries {

		s := fmt.Sprintf("(%d, %d, %v,", a.Place, a.Number, a.Tension) +
			"STRFTIME('%Y-%m-%d %H:%M:%f','" +
			a.Time.Format("2006-01-02 15:04:05.000") + "'))"
		if i < len(productVoltageSeries)-1 {
			s += ", "
		}
		queryStr += s
	}
	return queryStr
}

func queryInsertAmbient() string {
	queryStr := `INSERT INTO ambient(temperature, pressure, humidity, time) VALUES `
	for i, a := range ambientSeries {

		s := fmt.Sprintf("(%v, %v, %v,", a.Temperature, a.Pressure, a.Humidity) +
			"STRFTIME('%Y-%m-%d %H:%M:%f','" +
			a.Time.Format("2006-01-02 15:04:05.000") + "'))"
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

func init(){
	go func() {
		t := time.NewTicker(time.Minute)
		for {
			<-t.C
			SaveAndCleanCache()
		}
	}()
}
