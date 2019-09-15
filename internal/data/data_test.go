package data

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestCreateDB(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	x := OpenDev()
	newParty(x)
	for {
		var m Measurement
		for j := range m {
			v := randV()
			m[j] = &v
		}
		x.AddMeasurement(m)
		time.Sleep(2 * time.Millisecond)
		count := len(x.ms)

		if count >= 10000 {
			fmt.Println("saving")
			t := time.Now()
			x.Save()
			fmt.Println("save", count, time.Since(t))
		}
	}
}

func randV() float64 {
	return math.Floor(rand.Float64()*100) / 100
}

func newParty(x *DB) {
	r, err := x.db.Exec(`INSERT INTO party (created_at) VALUES (?)`, time.Now())
	if err != nil {
		panic(err)
	}
	partyID, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}
	for place := 0; place < 50; place++ {
		x.db.MustExec(
			`INSERT INTO product (party_id, serial, place, product_type) VALUES (?, ?, ?, ?)`,
			partyID, place+100, place, place+50)
	}
}
