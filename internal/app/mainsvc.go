package app

import (
	"context"
	"github.com/fpawel/oxygen73/internal"
	"github.com/fpawel/oxygen73/internal/data"
	"github.com/fpawel/oxygen73/internal/guiclient"
	"github.com/fpawel/oxygen73/internal/thriftgen/apitypes"
	"github.com/fpawel/oxygen73/internal/thriftgen/mainsvc"
	"github.com/jmoiron/sqlx"
	"math"
	"time"
)

type mainSvcHandler struct {
	db           *sqlx.DB
	interruptApp context.CancelFunc
}

var _ mainsvc.MainSvc = mainSvcHandler{}

func (x mainSvcHandler) OpenClient(ctx context.Context) error {
	log.Info("open client")
	go log.ErrIfFail(guiclient.Open)
	return nil
}

func (x mainSvcHandler) CloseClient(ctx context.Context) error {
	log.Info("close client")
	go func() {
		log.ErrIfFail(guiclient.Close)
		if !internal.DevMode {
			x.interruptApp()
		}
	}()
	return nil
}

func (x mainSvcHandler) ListMeasurements(ctx context.Context, timeFrom apitypes.TimeUnixMillis,
	timeTo apitypes.TimeUnixMillis) ([]*apitypes.Measurement, error) {
	var xs []measurement
	err := x.db.Select(&xs, `
SELECT stored_at, 
       temperature, pressure, humidity,
       place0, place1, place2, place3, place4, place5, place6, place7, place8, place9,
       place10, place11, place12, place13, place14, place15, place16, place17, place18, place19,
       place20, place21, place22, place23, place24, place25, place26, place27, place28, place29,
       place30, place31, place32, place33, place34, place35, place36, place37, place38, place39,
       place40, place41, place42, place43, place44, place45, place46, place47, place48, place49
FROM measurement1 WHERE tm BETWEEN julianday(?) AND julianday(?)`,
		unixMillisToTime(timeFrom), unixMillisToTime(timeTo))
	if err != nil {
		return nil, err
	}
	ms := make([]*apitypes.Measurement, len(xs))
	f := floatOrNan
	for i, x := range xs {
		t := parseTime(x.StoredAt)
		ms[i] = &apitypes.Measurement{
			StoredAt:    timeUnixMillis(t),
			Temperature: f(x.Temperature),
			Pressure:    f(x.Pressure),
			Humidity:    f(x.Humidity),
			Places: []float64{
				f(x.Place0), f(x.Place1), f(x.Place2), f(x.Place3), f(x.Place4), f(x.Place5), f(x.Place6), f(x.Place7), f(x.Place8), f(x.Place9),
				f(x.Place10), f(x.Place11), f(x.Place12), f(x.Place13), f(x.Place14), f(x.Place15), f(x.Place16), f(x.Place17), f(x.Place18), f(x.Place19),
				f(x.Place20), f(x.Place21), f(x.Place22), f(x.Place23), f(x.Place24), f(x.Place25), f(x.Place26), f(x.Place27), f(x.Place28), f(x.Place29),
				f(x.Place30), f(x.Place31), f(x.Place32), f(x.Place33), f(x.Place34), f(x.Place35), f(x.Place36), f(x.Place37), f(x.Place38), f(x.Place39),
				f(x.Place40), f(x.Place41), f(x.Place42), f(x.Place43), f(x.Place44), f(x.Place45), f(x.Place46), f(x.Place47), f(x.Place48), f(x.Place49),
			},
		}
	}
	return ms, nil
}

func (x mainSvcHandler) ListYearMonths(ctx context.Context) ([]*apitypes.YearMonth, error) {
	var xs []*apitypes.YearMonth
	if err := x.db.Select(&xs, `
SELECT DISTINCT year,  month
FROM measurement1 
ORDER BY year DESC, month DESC`); err != nil {
		return nil, err
	}
	if len(xs) == 0 {
		t := time.Now()
		xs = append(xs, &apitypes.YearMonth{
			Year:  int32(t.Year()),
			Month: int32(t.Month()),
		})
	}
	return xs, nil
}

func (x mainSvcHandler) GetParty(ctx context.Context, partyID int64) (*apitypes.Party, error) {
	p, err := data.GetParty(x.db, partyID)
	if err != nil {
		return nil, err
	}
	return &apitypes.Party{
		PartyID:   p.PartyID,
		CreatedAt: timeUnixMillis(p.CreatedAt),
	}, nil
}

func (x mainSvcHandler) ListBucketsOfYearMonth(ctx context.Context, year int32, month int32) ([]*apitypes.Bucket, error) {
	var xs []struct {
		BucketID       int64     `db:"bucket_id"`
		PartyID        int64     `db:"party_id"`
		PartyCreatedAt time.Time `db:"party_created_at"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
		IsLast         bool      `db:"is_last"`
	}
	if err := x.db.Select(&xs, `
SELECT bucket_id, party_id, party_created_at, created_at, updated_at, is_last FROM bucket1
WHERE year = ? 
  AND month = ?`, year, month); err != nil {
		return nil, err
	}
	r := make([]*apitypes.Bucket, len(xs))
	for i, x := range xs {
		r[i] = &apitypes.Bucket{
			BucketID:       x.BucketID,
			CreatedAt:      timeUnixMillis(x.CreatedAt),
			UpdatedAt:      timeUnixMillis(x.UpdatedAt),
			PartyID:        x.PartyID,
			IsLast:         x.IsLast,
			PartyCreatedAt: timeUnixMillis(x.PartyCreatedAt),
		}
	}
	return r, nil
}

func (x mainSvcHandler) ListProducts(ctx context.Context, partyID int64) ([]*apitypes.Product, error) {
	xs, err := data.ListProducts(x.db, partyID)
	if err != nil {
		return nil, err
	}
	ps := make([]*apitypes.Product, len(xs))
	for i, p := range xs {
		ps[i] = &apitypes.Product{
			Place:     p.Place,
			ProductID: p.ProductID,
			PartyID:   p.PartyID,
			Serial:    p.Serial,
		}
	}
	return ps, nil
}

func (x mainSvcHandler) CreateNewParty(ctx context.Context, products []*apitypes.Product) error {
	return nil
}

const timeLayout = "2006-01-02 15:04:05.000"

func timeUnixMillis(t time.Time) apitypes.TimeUnixMillis {

	return apitypes.TimeUnixMillis(t.UnixNano() / int64(time.Millisecond))
}

func unixMillisToTime(m apitypes.TimeUnixMillis) time.Time {
	t := int64(time.Millisecond) * int64(m)
	sec := t / int64(time.Second)
	ns := t % int64(time.Second)
	return time.Unix(sec, ns)
}

func floatOrNan(x *float64) float64 {
	if x == nil {
		return math.NaN()
	}
	return *x
}

func parseTime(sqlStr string) time.Time {
	t, err := time.ParseInLocation(timeLayout, sqlStr, time.Now().Location())
	if err != nil {
		panic(err)
	}
	return t
}

type measurement struct {
	StoredAt    string   `db:"stored_at"`
	Temperature *float64 `db:"temperature"`
	Pressure    *float64 `db:"pressure"`
	Humidity    *float64 `db:"humidity"`

	Place0 *float64 `db:"place0"`
	Place1 *float64 `db:"place1"`
	Place2 *float64 `db:"place2"`
	Place3 *float64 `db:"place3"`
	Place4 *float64 `db:"place4"`
	Place5 *float64 `db:"place5"`
	Place6 *float64 `db:"place6"`
	Place7 *float64 `db:"place7"`
	Place8 *float64 `db:"place8"`
	Place9 *float64 `db:"place9"`

	Place10 *float64 `db:"place10"`
	Place11 *float64 `db:"place11"`
	Place12 *float64 `db:"place12"`
	Place13 *float64 `db:"place13"`
	Place14 *float64 `db:"place14"`
	Place15 *float64 `db:"place15"`
	Place16 *float64 `db:"place16"`
	Place17 *float64 `db:"place17"`
	Place18 *float64 `db:"place18"`
	Place19 *float64 `db:"place19"`

	Place20 *float64 `db:"place20"`
	Place21 *float64 `db:"place21"`
	Place22 *float64 `db:"place22"`
	Place23 *float64 `db:"place23"`
	Place24 *float64 `db:"place24"`
	Place25 *float64 `db:"place25"`
	Place26 *float64 `db:"place26"`
	Place27 *float64 `db:"place27"`
	Place28 *float64 `db:"place28"`
	Place29 *float64 `db:"place29"`

	Place30 *float64 `db:"place30"`
	Place31 *float64 `db:"place31"`
	Place32 *float64 `db:"place32"`
	Place33 *float64 `db:"place33"`
	Place34 *float64 `db:"place34"`
	Place35 *float64 `db:"place35"`
	Place36 *float64 `db:"place36"`
	Place37 *float64 `db:"place37"`
	Place38 *float64 `db:"place38"`
	Place39 *float64 `db:"place39"`

	Place40 *float64 `db:"place40"`
	Place41 *float64 `db:"place41"`
	Place42 *float64 `db:"place42"`
	Place43 *float64 `db:"place43"`
	Place44 *float64 `db:"place44"`
	Place45 *float64 `db:"place45"`
	Place46 *float64 `db:"place46"`
	Place47 *float64 `db:"place47"`
	Place48 *float64 `db:"place48"`
	Place49 *float64 `db:"place49"`
}
