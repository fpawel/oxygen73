package data

import (
	"context"
	"database/sql"
	"github.com/fpawel/oxygen73/internal/pkg/must"
	"github.com/jmoiron/sqlx"
	"os"
	"path/filepath"
	"time"
)

//go:generate go run github.com/fpawel/gotools/cmd/sqlstr/...

type Product struct {
	Place          int32     `db:"place"`
	ProductID      int64     `db:"product_id"`
	PartyID        int64     `db:"party_id"`
	Serial         int32     `db:"serial"`
	PartyCreatedAt time.Time `db:"created_at"`
}

type Party struct {
	CreatedAt time.Time `db:"created_at"`
	PartyID   int64     `db:"party_id"`
}

func OpenDev() *sqlx.DB {
	return Open(filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "fpawel", "oxygen73", "build", "oxygen73.sqlite"))
}

func OpenProd() *sqlx.DB {
	return Open(filepath.Join(filepath.Dir(os.Args[0]), "oxygen73.sqlite"))
}

func Open(filename string) *sqlx.DB {
	db := must.OpenSqliteDBx(filename)
	db.MustExec(SQLCreate)

	// создать новую партию, если партия не была

	return db
}

func GetParty(ctx context.Context, db *sqlx.DB, partyID int64) (p Party, err error) {
	if partyID < 0 {
		err = db.GetContext(ctx, &p, `SELECT * FROM last_party`)
	} else {
		err = db.GetContext(ctx, &p, `SELECT * FROM  party WHERE party_id = ?`, partyID)
	}
	return
}

func ListProducts(ctx context.Context, db *sqlx.DB, partyID int64) ([]Product, error) {
	xs := make([]Product, 50)
	var (
		ps  []Product
		err error
	)
	err = db.SelectContext(ctx, &ps, `
SELECT place, product.product_id, product.party_id, serial, created_at   
FROM  product 
INNER JOIN party USING (party_id)
WHERE party_id = ? 
ORDER BY place`, partyID)
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

func LastParty(ctx context.Context, db *sqlx.DB) (party Party, err error) {
	err = db.GetContext(ctx, &party, `SELECT * FROM last_party`)
	if err == nil {
		return
	}
	if err != sql.ErrNoRows {
		return
	}

	if _, err = db.ExecContext(ctx, `INSERT INTO party DEFAULT VALUES;`); err != nil {
		return
	}
	if err = db.GetContext(ctx, &party, `SELECT * FROM last_party`); err != nil {
		return
	}
	_, err = db.ExecContext(ctx, `INSERT INTO product(party_id, serial, place) VALUES (?, 1, 0);`, party.PartyID)
	return
}
