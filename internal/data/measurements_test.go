package data

import (
	"github.com/jmoiron/sqlx"
	"log"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestCreateDB(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	db := OpenDev()
	newParty(db)
	var mss Measurements
	for {
		var m Measurement
		for j := range m.Places {
			m.Places[j] = randV()
		}
		m.Humidity = randV()
		m.Pressure = randV()
		m.Temperature = randV()
		m.StoredAt = time.Now()
		mss = append(mss, m)

		time.Sleep(time.Second)
		count := len(mss)

		if count >= 10 {
			log.Println("saving")
			t := time.Now()
			log.Println("save", count, time.Since(t), SaveMeasurements(mss, db))
			mss = nil
		}
	}
}

func randV() float64 {
	return math.Floor(rand.Float64()*100) / 100
}

func newParty(db *sqlx.DB) {
	r, err := db.Exec(`INSERT INTO party (created_at) VALUES (?)`, time.Now())
	if err != nil {
		panic(err)
	}
	partyID, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}
	for place := 0; place < 50; place++ {
		m := int(time.Now().Month())
		y := time.Now().Year()
		db.MustExec(
			`
INSERT INTO product (party_id, serial, place  ) 
VALUES (?, ?, ? )`,
			partyID, place+100+y*12+m, place)
	}
}
