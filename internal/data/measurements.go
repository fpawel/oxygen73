package data

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/powerman/structlog"
	"strconv"
	"strings"
	"time"
)

type Measurements = []Measurement

type Measurement struct {
	Temperature,
	Pressure,
	Humidity float64
	Places   [50]float64
	StoredAt time.Time
}

func SaveMeasurements(measurements Measurements, db *sqlx.DB) error {
	var xs []string
	for _, m := range measurements {
		var xsv []string
		for _, v := range m.Places {
			var s string
			if v < 60 {
				s = strconv.FormatFloat(v, 'f', -1, 64)
			} else {
				s = "NULL"
			}
			xsv = append(xsv, fmt.Sprintf("%v", s))
		}
		xsv = append(xsv, fmt.Sprintf("%v,%v,%v", m.Temperature, m.Pressure, m.Humidity))
		xs = append(xs, fmt.Sprintf("(%s, %s)", formatTimeAsQuery(m.StoredAt), strings.Join(xsv, ",")))
	}

	strQueryInsert := `INSERT INTO measurement(
tm, 
place0, place1, place2, place3, place4, place5, place6, place7, place8, place9, 
place10, place11, place12, place13, place14, place15, place16, place17, place18, place19,
place20, place21, place22, place23, place24, place25, place26, place27, place28, place29, 
place30, place31, place32, place33, place34, place35, place36, place37, place38, place39, 
place40, place41, place42, place43, place44, place45, place46, place47, place48, place49, temperature, pressure, humidity
) VALUES ` + "  " + strings.Join(xs, ",")
	if _, err := db.Exec(strQueryInsert); err != nil {
		err = fmt.Errorf("fail to insert measurements: %w", err)
		return log.Err(err, "sql", fmt.Sprintf("`%s`", strQueryInsert))
	}
	return nil
}

//func parseTime(sqlStr string) time.Time {
//	t, err := time.ParseInLocation(timeLayout, sqlStr, time.Now().Location())
//	if err != nil {
//		panic(err)
//	}
//	return t
//}

func formatTimeAsQuery(t time.Time) string {
	return "julianday(STRFTIME('%Y-%m-%d %H:%M:%f','" +
		t.Format(timeLayout) + "'))"
}

const timeLayout = "2006-01-02 15:04:05.000"

var log = structlog.New()
