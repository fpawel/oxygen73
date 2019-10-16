package data

import (
	"database/sql"
	"github.com/fpawel/oxygen73/internal/pkg/must"
	"github.com/jmoiron/sqlx"
	"os"
	"path/filepath"
	"time"
)

//go:generate go run github.com/fpawel/gotools/cmd/sqlstr/...

type Product struct {
	Place     int32 `db:"place"`
	ProductID int64 `db:"product_id"`
	PartyID   int64 `db:"party_id"`
	Serial    int32 `db:"serial"`
}

type Party struct {
	CreatedAt time.Time `db:"created_at"`
	PartyID   int64     `db:"party_id"`
}

func OpenDev() *sqlx.DB {
	return Open(filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "fpawel", "oxygen73", "build", "series.sqlite"))
}

func OpenProd() *sqlx.DB {
	return Open(filepath.Join(filepath.Dir(os.Args[0]), "series.sqlite"))
}

func Open(filename string) *sqlx.DB {
	db := must.OpenSqliteDBx(filename)
	db.MustExec(SQLCreate)

	// создать новую партию, если партия не была

	return db
}

func GetParty(db *sqlx.DB, partyID int64) (p Party, err error) {
	if partyID < 0 {
		err = db.Get(&p, `SELECT * FROM last_party`)
	} else {
		err = db.Get(&p, `SELECT * FROM  party WHERE party_id = ?`, partyID)
	}
	return
}

func ListProducts(db *sqlx.DB, partyID int64) ([]Product, error) {
	xs := make([]Product, 50)
	var (
		ps  []Product
		err error
	)
	if partyID < 0 {
		err = db.Select(&ps, `
SELECT * 
FROM  product 
WHERE party_id = (SELECT party_id FROM last_party) 
ORDER BY place`)
	} else {
		err = db.Select(&ps, ` SELECT * FROM  product WHERE party_id = ? ORDER BY place`, partyID)
	}
	if err != nil {
		return nil, err
	}
	for _, p := range ps {
		if p.Place < 0 || int(p.Place) >= len(xs) {
			panic("unexpected")
		}
		xs[p.Place] = p
	}
	return xs, nil
}

func MustLastParty(db *sqlx.DB) (party Party) {
	err := db.Get(&party, `SELECT * FROM last_party`)
	if err == nil {
		return
	}
	if err != sql.ErrNoRows {
		panic(err)
	}
	db.MustExec(`INSERT INTO party DEFAULT VALUES;`)
	if err = db.Get(&party, `SELECT * FROM last_party`); err != nil {
		panic(err)
	}
	db.MustExec(`INSERT INTO product(party_id, serial, place) VALUES (?, 1, 0);`, party.PartyID)
	return
}
