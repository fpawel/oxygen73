package data

import (
	"fmt"
	"github.com/fpawel/gohelp"
	"github.com/fpawel/oxygen73/internal"
	"github.com/jmoiron/sqlx"
	"path/filepath"
	"sync"
	"time"
)

//go:generate go run github.com/fpawel/gohelp/cmd/sqlstr/...

func AddProductVoltage(place int, voltage float64) {
	mu.Lock()
	defer mu.Unlock()
	productVoltageSeries = append(productVoltageSeries, productVoltageSample{
		StoredAt: time.Now(),
		Place:    place,
		Voltage:  voltage,
	})
}

func AddAmbient(temperature, pressure, humidity float64) {
	mu.Lock()
	defer mu.Unlock()
	ambientSeries = append(ambientSeries, ambientSample{
		StoredAt:    time.Now(),
		Temperature: temperature,
		Pressure:    pressure,
		Humidity:    humidity,
	})
}

func Save() {
	mu.Lock()
	defer mu.Unlock()
	save()
}

func save() {
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
	StoredAt time.Time
	Place    int
	Voltage  float64
}

type ambientSample struct {
	StoredAt    time.Time
	Temperature float64
	Pressure    float64
	Humidity    float64
}

func queryInsertProductVoltages() string {
	m := make(map[int]int64)

	for _, x := range productVoltageSeries {
		if _, f := m[x.Place]; f {
			continue
		}
		var productID int
		err := db.Get(productID,
			`SELECT product_id FROM product WHERE party_id = (SELECT party_id FROM last_party) AND place = ?`,
			x.Place)
	}

	queryStr := `INSERT INTO product_voltage(stored_at, product_id, voltage) VALUES `
	for i, a := range productVoltageSeries {

		s := "(" + formatTimeAsQuery(a.StoredAt) + fmt.Sprintf(", %d, %v)", a.ProductID, a.Voltage)
		if i < len(productVoltageSeries)-1 {
			s += ", "
		}
		queryStr += s
	}
	return queryStr
}

func queryInsertAmbient() string {
	queryStr := `INSERT INTO ambient(stored_at, temperature, pressure, humidity, ) VALUES `
	for i, a := range ambientSeries {
		s := "(" + formatTimeAsQuery(a.StoredAt) + "," +
			fmt.Sprintf("%v, %v, %v)", a.Temperature, a.Pressure, a.Humidity)
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
