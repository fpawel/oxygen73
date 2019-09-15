// создаёт тестовую базу данных oxygen73lab в локальной ноде influxdb
package main

import (
	"fmt"
	"github.com/ansel1/merry"
	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/powerman/structlog"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const (
	DataBaseName = "oxygen73lab"
)

func main() {
	c, err := influx.NewHTTPClient(influx.HTTPConfig{Addr: "http://localhost:8086"})
	if err != nil {
		log.Fatal("Error creating InfluxDB Client:", err)
	}
	defer log.ErrIfFail(c.Close)

	influxMustExecQuery(influx.NewQuery("DROP DATABASE "+DataBaseName, "", ""), c)
	influxMustExecQuery(influx.NewQuery("CREATE DATABASE "+DataBaseName, "", ""), c)
	influxMustExecQuery(influx.NewQuery("USE "+DataBaseName, "", ""), c)
	if !influxIsDatabaseExists(DataBaseName,
		influxMustExecQuery(influx.NewQuery("SHOW DATABASES", "", ""), c)) {
		panic("expect database was created")
	}

	for d := 1; d < 5; d++ {
		for h := 8; h <= 12; h++ {
			for m := 0; m <= 59; m++ {
				bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{Database: DataBaseName})
				if err != nil {
					log.Fatal(err)
				}
				for s := 0; s <= 59; s++ {
					for _, ms := range []int{10, 555} {

						if err != nil {
							log.Fatal(err)
						}
						t := time.Date(2019, time.January, d, h, m, s, ms*1_000_000, time.Local)
						for place := 0; place < 50; place++ {
							bp.AddPoint(influxPoint("interrogate", map[string]string{
								"place":          fmt.Sprintf("%d", place),
								"product_serial": fmt.Sprintf("%d", 100*d+place),
							}, map[string]interface{}{
								"tension": rand.Float64() * 100,
							}, t))
						}
						bp.AddPoint(influxPoint("interrogate", nil, map[string]interface{}{
							"temperature": 20 + rand.Float64()*10,
							"humidity":    90 + rand.Float64()*10,
							"pressure":    rand.Float64() * 10,
						}, t))
					}
				}
				log.Println(d, h, m)
				if err := c.Write(bp); err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	log.Println("done")
}

func influxPoint(name string, tags map[string]string, fields map[string]interface{}, t time.Time) *influx.Point {
	p, err := influx.NewPoint(name, tags, fields, t)
	if err != nil {
		log.Fatal(err)
	}
	return p
}

func influxMustExecQuery(q influx.Query, c influx.Client) *influx.Response {
	r, err := influxExecQuery(q, c)
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func influxExecQuery(q influx.Query, c influx.Client) (*influx.Response, error) {
	response, err := c.Query(q)
	if err != nil {
		return nil, merry.Append(err, "executing query")
	}
	if response.Error() != nil {
		return nil, merry.Append(err, "bad response")
	}
	return response, nil
}

func influxIsDatabaseExists(dataBaseName string, response *influx.Response) bool {
	for _, r := range response.Results {
		for _, m := range r.Messages {
			log.Info(m.Text, "influx_message_level")
		}
		for _, row := range r.Series {
			for _, v := range row.Values {
				if v[0].(string) == dataBaseName {
					return true
				}
			}
		}
	}
	return false
}

var (
	log = structlog.New()
)

func init() {
	rand.Seed(time.Now().UnixNano())
	structlog.DefaultLogger.
		SetPrefixKeys(structlog.KeyApp, structlog.KeyPID, structlog.KeyLevel, structlog.KeyUnit).
		SetSuffixKeys(structlog.KeySource, structlog.KeyStack).
		SetDefaultKeyvals(
			structlog.KeyApp, filepath.Base(os.Args[0]),
			structlog.KeySource, structlog.Auto,
		).
		SetKeysFormat(map[string]string{
			structlog.KeySource: " %6[2]s",
			structlog.KeyUnit:   " %6[2]s",
		})
}
