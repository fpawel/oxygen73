package app

import (
	"context"
	"fmt"
	"github.com/fpawel/gotools/pkg/logfile"
	"github.com/fpawel/oxygen73/internal/cfg"
	"github.com/fpawel/oxygen73/internal/data"
	"github.com/fpawel/oxygen73/internal/gui"
	"github.com/fpawel/oxygen73/internal/thriftgen/apitypes"
	"github.com/fpawel/oxygen73/internal/thriftgen/mainsvc"
	"github.com/jmoiron/sqlx"
	"github.com/lxn/win"
	"math"
	"time"
)

type mainSvcHandler struct {
	db *sqlx.DB
}

var _ mainsvc.MainSvc = new(mainSvcHandler)

func (x *mainSvcHandler) DeleteBucket(ctx context.Context, bucketID int64) error {
	_, err := x.db.Exec(`DELETE FROM bucket WHERE bucket_id = ?`, bucketID)
	return err
}

func (x *mainSvcHandler) Vacuum(_ context.Context) error {
	go func() {
		t := time.Now()
		log.Info("vacuum begin: " + t.String())
		gui.VacuumBegin()
		_, err := x.db.Exec(`VACUUM`)
		gui.VacuumEnd()
		if err != nil {
			log.PrintErr(fmt.Sprintf("vacuum end: %v: %v: %v", t, time.Since(t), err))
			gui.ErrorOccurred(err)
			return
		}
		log.Info(fmt.Sprintf("vacuum end: %v: %v", t, time.Since(t)))
		gui.MsgBox("Данные ЭХЯ О2",
			fmt.Sprintf("Дефрагментация завершена успешно за %s", time.Since(t)),
			win.MB_ICONINFORMATION|win.MB_OK)

	}()
	return nil
}

func (x *mainSvcHandler) GetAppConfigYaml(_ context.Context) (string, error) {
	return cfg.GetYaml(), nil
}

func (x *mainSvcHandler) SetAppConfigYaml(_ context.Context, appConfigToml string) error {
	return cfg.SetYaml(appConfigToml)
}

func (x *mainSvcHandler) GetAppConfig(_ context.Context) (*apitypes.AppConfig, error) {
	c := cfg.Get()
	return &apitypes.AppConfig{
		Comport:         c.Main.Comport,
		ComportHumidity: c.Hum.Comport,
	}, nil
}

func (x *mainSvcHandler) SetAppConfig(_ context.Context, appConfig *apitypes.AppConfig) error {
	c := cfg.Get()
	c.Main.Comport = appConfig.Comport
	c.Hum.Comport = appConfig.ComportHumidity
	cfg.Set(c)
	return nil
}

func (x *mainSvcHandler) ListLastPartyProducts(ctx context.Context) ([]*apitypes.Product, error) {
	var partyID int64
	if err := x.db.GetContext(ctx, &partyID, `SELECT party_id FROM last_party`); err != nil {
		return nil, err
	}
	return x.ListProducts(ctx, partyID)
}

func (x *mainSvcHandler) SetProductSerialAtPlace(ctx context.Context, place int32, serial int32) (err error) {
	_, err = x.db.ExecContext(ctx, `
INSERT OR REPLACE INTO product(party_id, serial, place)
VALUES ((SELECT party_id FROM last_party), ?, ?)
`, serial, place)
	return
}

func (x *mainSvcHandler) DeleteProductAtPlace(ctx context.Context, place int32) (err error) {
	_, err = x.db.ExecContext(ctx, `
DELETE FROM product WHERE party_id=(SELECT party_id FROM last_party) AND place=?`, place)
	return
}

func (x *mainSvcHandler) RequestProductMeasurements(ctx context.Context, bucketID int64, place int32) error {
	go func() {
		var xs []struct {
			StoredAt    string   `db:"stored_at"`
			Temperature *float64 `db:"temperature"`
			Pressure    *float64 `db:"pressure"`
			Humidity    *float64 `db:"humidity"`
			Value       float64  `db:"value"`
		}

		sqlStr := fmt.Sprintf(`
SELECT stored_at, 
       temperature, pressure, humidity, place%d AS value
FROM measurement1 
WHERE tm BETWEEN julianday((SELECT created_at FROM bucket WHERE bucket_id=?)) AND julianday((SELECT updated_at FROM bucket WHERE bucket_id=?))
	AND place%d NOT NULL`, place, place)

		err := x.db.SelectContext(ctx, &xs, sqlStr, bucketID, bucketID)
		if err != nil {
			log.PrintErr("select measurements fail",
				"reason", err,
				"bucketID", bucketID,
				"place", place)
			gui.ErrorOccurred(err)
			return
		}

		ms := make([]data.Measurement1, len(xs))
		f := floatOrNan
		for i, x := range xs {
			t := parseTime(x.StoredAt)
			ms[i] = data.Measurement1{
				StoredAt:    t,
				Temperature: f(x.Temperature),
				Pressure:    f(x.Pressure),
				Humidity:    f(x.Humidity),
				Value:       x.Value,
			}
		}
		gui.ProductMeasurements(bucketID, ms)
	}()
	return nil
}

func (x *mainSvcHandler) RequestMeasurements(ctx context.Context, bucketID int64) error {
	go func() {

		log.Debug(fmt.Sprintf("opening bucket %d", bucketID))

		var xs []measurement
		t := time.Now()
		err := x.db.SelectContext(ctx, &xs, `
SELECT stored_at, 
       temperature, pressure, humidity,
       place0, place1, place2, place3, place4, place5, place6, place7, place8, place9,
       place10, place11, place12, place13, place14, place15, place16, place17, place18, place19,
       place20, place21, place22, place23, place24, place25, place26, place27, place28, place29,
       place30, place31, place32, place33, place34, place35, place36, place37, place38, place39,
       place40, place41, place42, place43, place44, place45, place46, place47, place48, place49
FROM measurement1 WHERE tm BETWEEN
    julianday((SELECT created_at FROM bucket WHERE bucket_id = ?)) AND 
    julianday((SELECT updated_at FROM bucket WHERE bucket_id = ?))`, bucketID, bucketID)
		if err != nil {
			log.PrintErr("select measurements fail",
				"reason", err,
				"bucket_id", bucketID)
			return
		}
		ms := make([]data.Measurement, len(xs))
		f := floatOrNan
		for i, x := range xs {
			t := parseTime(x.StoredAt)
			ms[i] = data.Measurement{
				StoredAt:    t,
				Temperature: f(x.Temperature),
				Pressure:    f(x.Pressure),
				Humidity:    f(x.Humidity),
				Places: [50]float64{
					f(x.Place0), f(x.Place1), f(x.Place2), f(x.Place3), f(x.Place4), f(x.Place5), f(x.Place6), f(x.Place7), f(x.Place8), f(x.Place9),
					f(x.Place10), f(x.Place11), f(x.Place12), f(x.Place13), f(x.Place14), f(x.Place15), f(x.Place16), f(x.Place17), f(x.Place18), f(x.Place19),
					f(x.Place20), f(x.Place21), f(x.Place22), f(x.Place23), f(x.Place24), f(x.Place25), f(x.Place26), f(x.Place27), f(x.Place28), f(x.Place29),
					f(x.Place30), f(x.Place31), f(x.Place32), f(x.Place33), f(x.Place34), f(x.Place35), f(x.Place36), f(x.Place37), f(x.Place38), f(x.Place39),
					f(x.Place40), f(x.Place41), f(x.Place42), f(x.Place43), f(x.Place44), f(x.Place45), f(x.Place46), f(x.Place47), f(x.Place48), f(x.Place49),
				},
			}
		}
		log.Debug(fmt.Sprintf("bucket %d: %d measurements: %v", bucketID, len(ms), time.Since(t)))
		gui.Measurements(bucketID, ms)
	}()
	return nil
}

func (x *mainSvcHandler) ListYearMonths(ctx context.Context) ([]*apitypes.YearMonth, error) {
	var xs []*apitypes.YearMonth
	if err := x.db.SelectContext(ctx, &xs, `
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

func (x *mainSvcHandler) GetParty(ctx context.Context, partyID int64) (*apitypes.Party, error) {
	p, err := data.GetParty(ctx, x.db, partyID)
	if err != nil {
		return nil, err
	}
	return &apitypes.Party{
		PartyID:   p.PartyID,
		CreatedAt: timeUnixMillis(p.CreatedAt),
	}, nil
}

func (x *mainSvcHandler) ListBucketsOfYearMonth(_ context.Context, year int32, month int32) ([]*apitypes.Bucket, error) {
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

func (x *mainSvcHandler) ListProducts(ctx context.Context, partyID int64) ([]*apitypes.Product, error) {
	xs, err := data.ListProducts(ctx, x.db, partyID)
	if err != nil {
		return nil, err
	}
	ps := make([]*apitypes.Product, len(xs))
	for i, p := range xs {
		ps[i] = &apitypes.Product{
			Place:          p.Place,
			ProductID:      p.ProductID,
			PartyID:        p.PartyID,
			Serial:         p.Serial,
			PartyCreatedAt: timeUnixMillis(p.PartyCreatedAt),
		}
	}
	return ps, nil
}

func (x *mainSvcHandler) CreateNewParty(ctx context.Context) error {
	r, err := x.db.ExecContext(ctx, `INSERT INTO party DEFAULT VALUES`)
	if err != nil {
		return err
	}
	if _, err := r.LastInsertId(); err != nil {
		return err
	}
	return nil
}

func (x *mainSvcHandler) ListLogEntriesDays(_ context.Context) (r []apitypes.TimeUnixMillis, _ error) {
	for _, t := range logfile.ListDays() {
		r = append(r, timeUnixMillis(t))
	}
	return
}

func (x *mainSvcHandler) LogEntriesOfDay(_ context.Context, daytime apitypes.TimeUnixMillis, filter string) (r []*apitypes.LogEntry, err error) {

	var xs []logfile.Entry
	t := unixMillisToTime(daytime)
	t = t.In(time.Local)

	if xs, err = logfile.Read(t, filter); err != nil {
		return nil, err
	}

	for _, a := range xs {
		r = append(r, &apitypes.LogEntry{
			Time: timeUnixMillis(a.Time),
			Line: a.Line,
		})
	}
	return
}

func (x *mainSvcHandler) FindProductsBySerial(ctx context.Context, serial int32) ([]*apitypes.ProductBucket, error) {
	var xs []struct {
		Place           int32     `db:"place"`
		ProductID       int64     `db:"product_id"`
		PartyID         int64     `db:"party_id"`
		BucketID        int64     `db:"bucket_id"`
		Serial          int32     `db:"serial"`
		PartyCreatedAt  time.Time `db:"party_created_at"`
		BucketCreatedAt time.Time `db:"bucket_created_at"`
		BucketUpdatedAt time.Time `db:"bucket_updated_at"`
	}
	err := x.db.SelectContext(ctx, &xs, `
SELECT product.product_id AS product_id,
       place,
       serial,
       party.party_id AS party_id,
       party.created_at  AS party_created_at,
       bucket.bucket_id AS bucket_id,
       bucket.created_at AS bucket_created_at,
       bucket.updated_at AS bucket_updated_at
FROM product
         INNER JOIN party USING (party_id)
         INNER JOIN bucket USING (party_id)
WHERE serial = ?
ORDER BY bucket.created_at; `, serial)
	if err != nil {
		return nil, err
	}
	r := make([]*apitypes.ProductBucket, 0)
	for _, p := range xs {

		r = append(r, &apitypes.ProductBucket{
			Place:           p.Place,
			ProductID:       p.ProductID,
			PartyID:         p.PartyID,
			Serial:          p.Serial,
			PartyCreatedAt:  timeUnixMillis(p.PartyCreatedAt),
			BucketID:        p.BucketID,
			BucketCreatedAt: timeUnixMillis(p.BucketCreatedAt),
			BucketUpdatedAt: timeUnixMillis(p.BucketUpdatedAt),
		})
	}
	return r, nil
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
	Tm          float64  `db:"tm"`
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
