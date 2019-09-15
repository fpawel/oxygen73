package data

import (
	"fmt"
	"github.com/fpawel/gohelp"
	"github.com/jmoiron/sqlx"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//go:generate go run github.com/fpawel/gotools/cmd/sqlstr/...

type DB struct {
	db *sqlx.DB
	ms []measurement
	mu sync.Mutex
}

type Measurement [53]*float64

func OpenDev() *DB {
	return Open(filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "fpawel", "oxygen73", "build", "series.sqlite"))
}

func OpenProd() *DB {
	return Open(filepath.Join(filepath.Dir(os.Args[0]), "series.sqlite"))
}

func Open(filename string) *DB {
	db := gohelp.MustOpenSqliteDBx(filename)
	db.MustExec(SQLCreate)
	return &DB{db: db}
}

func (x *DB) AddMeasurement(m Measurement) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.ms = append(x.ms, measurement{
		Measurement: m,
		StoredAt:    time.Now(),
	})
}

func (x *DB) Save() {
	x.mu.Lock()
	queryInsertMeasurements := x.queryInsertMeasurements()
	x.mu.Unlock()
	if len(queryInsertMeasurements) > 0 {
		x.db.MustExec(queryInsertMeasurements)
	}
}

type measurement struct {
	Measurement
	StoredAt time.Time
}

func (x *DB) queryInsertMeasurements() string {
	if len(x.ms) == 0 {
		return ""
	}
	var xs []string
	for _, m := range x.ms {
		var xsv []string
		for _, v := range m.Measurement {
			xsv = append(xsv, fmt.Sprintf("%v", v))
		}
		xs = append(xs, fmt.Sprintf("(%s, %s)", formatTimeAsQuery(m.StoredAt), strings.Join(xsv, ",")))
	}
	x.ms = nil
	return `INSERT INTO measurement(
stored_at, 
place0, place1, place2, place3, place4, place5, place6, place7, place8, place9, 
place10, place11, place12, place13, place14, place15, place16, place17, place18, place19,
place20, place21, place22, place23, place24, place25, place26, place27, place28, place29, 
place30, place31, place32, place33, place34, place35, place36, place37, place38, place39, 
place40, place41, place42, place43, place44, place45, place46, place47, place48, place49, temperature, pressure, humidity
) VALUES ` + strings.Join(xs, ",")
}

func parseTime(sqlStr string) time.Time {
	t, err := time.ParseInLocation(timeLayout, sqlStr, time.Now().Location())
	if err != nil {
		panic(err)
	}
	return t
}
func formatTimeAsQuery(t time.Time) string {
	return "julianday(STRFTIME('%Y-%m-%d %H:%M:%f','" +
		t.Format(timeLayout) + "'))"
}

const timeLayout = "2006-01-02 15:04:05.000"
