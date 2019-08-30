package data

import (
	"encoding/binary"
	"fmt"
	"github.com/powerman/must"
	"github.com/powerman/structlog"
	"io"
	"net/http"
	"strconv"
	"time"
)

type ChartsSvc struct{}

type YearMonth struct {
	Year  int `db:"created_at_year"`
	Month int `db:"created_at_month"`
}

type Series struct {
	StartedAtJulianDay float64
	StartedAt          TimeDelphi
	UpdatedAt          TimeDelphi
	IsLast             bool
}

type TimeDelphi struct {
	Year        int
	Month       time.Month
	Day         int
	Hour        int
	Minute      int
	Second      int
	Millisecond int
}

func (_ *ChartsSvc) YearsMonths(_ struct{}, r *[]YearMonth) error {
	if err := db.Select(r, `
SELECT DISTINCT year, month 
FROM product_voltage_series 
ORDER BY started_at DESC`); err != nil {
		panic(err)
	}
	return nil
}

func (_ *ChartsSvc) SeriesOfYearMonth(x YearMonth, r *[]Series) error {

	var xs []struct {
		StartedAtJulianDay float64 `db:"started_at"`
		StartedAt          string  `db:"started_at_str"`
		UpdatedAt          string  `db:"updated_at_str"`
		IsLast             bool    `db:"is_last"`
	}
	if err := db.Select(&xs, `
SELECT * FROM product_voltage_series
WHERE year = ?
  AND month = ?
ORDER BY started_at`, x.Year, x.Month); err != nil {
		panic(err)
	}
	for _, x := range xs {
		*r = append(*r, Series{
			StartedAt: timeDelphi(parseTime(x.StartedAt)),
			UpdatedAt: timeDelphi(parseTime(x.UpdatedAt)),
			IsLast:    x.IsLast,
		})
	}
	return nil
}

func (_ *ChartsSvc) DeletePoints(r struct {
	TimeFrom, TimeTo TimeDelphi
}, rowsAffected *int64) error {

	mu.Lock()
	n := 0
	for _, x := range productVoltageSeries {
		if x.StoredAt.After(r.TimeFrom.Time()) && x.StoredAt.Before(r.TimeTo.Time()) {
			productVoltageSeries[n] = x
			n++
		}
	}
	productVoltageSeries = productVoltageSeries[:n]
	n = 0
	for _, x := range ambientSeries {
		if x.StoredAt.After(r.TimeFrom.Time()) && x.StoredAt.Before(r.TimeTo.Time()) {
			ambientSeries[n] = x
			n++
		}
	}
	ambientSeries = ambientSeries[:n]
	mu.Unlock()

	var err error
	*rowsAffected, err = db.MustExec(
		`
DELETE FROM product_voltage 
WHERE stored_at >= julianday(?) AND stored_at <= julianday(?) `,
		r.TimeFrom.Time().Format(parseTimeFormat),
		r.TimeTo.Time().Format(parseTimeFormat)).RowsAffected()
	return err
}

func HandleRequestSeries(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept", "application/octet-stream")
	qStr := r.URL.Query().Get("series")
	series, err := strconv.ParseFloat(qStr, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		must.Write(w, []byte( fmt.Sprintf("series: %q: %v", qStr, err) ))
		return
	}
	type ptDelphi = struct {
		TimeDelphi
		Place        int     `db:"place"`
		SerialNumber int     `db:"serial_number"`
		Tension      float64 `db:"tension"`
	}
	var pts []ptDelphi

	mu.Lock()
	for _, x := range productVoltageSeries {
		pts = append(pts, ptDelphi{
			TimeDelphi:   timeDelphi(x.StoredAt),
			Place:        x.Place,
			SerialNumber: x.SerialNumber,
			Tension:      x.Tension,
		})
	}
	mu.Unlock()

	type ptSql = struct {
		StoredAt string  `db:"started_at_str"`
		Place    int     `db:"place"`
		Serial   int     `db:"serial_number"`
		Tension  float64 `db:"tension"`
	}
	var xs []ptSql

	if err := db.Select(&xs, `
SELECT stored_at_str, place, serial_number, tension 
FROM product_voltage_time 
WHERE series_created_at = ?`, series; err != nil {
		panic(err)
	}
	for _,x := range xs{
		pts = append(pts, ptDelphi{
			TimeDelphi:   timeDelphi(parseTime(x.StoredAt)),
			Place:        x.Place,
			SerialNumber: x.Serial,
			Tension:      x.Tension,
		})
	}

	write := func(n interface{}) {
		if err := binary.Write(w, binary.LittleEndian, n); err != nil {
			panic(err)
		}
	}
	write(uint64(len(points)))
	for _, x := range points {
		write(byte(x.Addr))
		write(uint16(x.Var))
		write(uint16(x.Year))
		write(byte(x.Month))
		write(byte(x.Day))
		write(byte(x.Hour))
		write(byte(x.Minute))
		write(byte(x.Second))
		write(uint16(x.Millisecond))
		write(x.Value)
	}
}

func writePointsResponse(w io.Writer, bucketID int64) {

	var points []point3

	if err := db.Select(&points, `
SELECT stored_at_str, place, serial_number, tension 
FROM product_voltage_time 
WHERE bucket_id = ?`, bucketID); err != nil {
		panic(err)
	}

	if b, f := lastBucket(); f && b.BucketID == bucketID {
		var points3 []point3
		muPoints.Lock()
		for _, p := range currentPoints {
			points3 = append(points3, p.point3())
		}
		muPoints.Unlock()
		points = append(points3, points...)
	}

	write := func(n interface{}) {
		if err := binary.Write(w, binary.LittleEndian, n); err != nil {
			panic(err)
		}
	}
	write(uint64(len(points)))
	for _, x := range points {
		write(byte(x.Addr))
		write(uint16(x.Var))
		write(uint16(x.Year))
		write(byte(x.Month))
		write(byte(x.Day))
		write(byte(x.Hour))
		write(byte(x.Minute))
		write(byte(x.Second))
		write(uint16(x.Millisecond))
		write(x.Value)
	}
}

func timeDelphi(t time.Time) TimeDelphi {
	return TimeDelphi{
		Year:        t.Year(),
		Month:       t.Month(),
		Day:         t.Day(),
		Hour:        t.Hour(),
		Minute:      t.Minute(),
		Second:      t.Second(),
		Millisecond: t.Nanosecond() / 1000000,
	}
}

func (x TimeDelphi) Time() time.Time {
	return time.Date(
		x.Year, x.Month, x.Day,
		x.Hour, x.Minute, x.Second,
		x.Millisecond*int(time.Millisecond/time.Nanosecond),
		time.Local)
}
