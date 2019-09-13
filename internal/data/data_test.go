package data

import (
	"math/rand"
	"testing"
	"time"
)

func TestCreateDB(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	tm := time.Now()

	r, err := db.Exec(`INSERT INTO party (created_at) VALUES (?)`, tm)
	if err != nil {
		panic(err)
	}
	partyID, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}
	for place := 0; place < 50; place++ {
		r, err := db.Exec(
			`INSERT INTO product (party_id, serial, place, product_type) VALUES (?, ?, ?, ?)`,
			partyID, place+100, place, place+50)
		if err != nil {
			panic(err)
		}
		productID, err := r.LastInsertId()
		if err != nil {
			panic(err)
		}

		for n := 0; n < 100; n++ {
			time.Sleep(2 * time.Millisecond)
			AddProductVoltage(productID, rand.Float64()*100)
		}
		AddAmbient(rand.Float64()*100, rand.Float64()*100, rand.Float64()*100)
		Save()
	}
}
