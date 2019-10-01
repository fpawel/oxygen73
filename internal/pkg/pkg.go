package pkg

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"unicode/utf16"
)

func UTF16FromString(s string) (b []byte) {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			panic(fmt.Sprintf("%q[%d] is 0", s, i))
		}
	}
	for _, v := range utf16.Encode([]rune(s)) {
		b = append(b, byte(v), byte(v>>8))
	}
	return
}

func OpenSqliteDBx(fileName string) (*sqlx.DB, error) {
	c, err := OpenSqliteDB(fileName)
	if err != nil {
		return nil, err
	}
	return sqlx.NewDb(c, "sqlite3"), nil
}

func OpenSqliteDB(fileName string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return nil, err
	}
	conn.SetMaxIdleConns(1)
	conn.SetMaxOpenConns(1)
	conn.SetConnMaxLifetime(0)
	return conn, err
}

func EnsuredDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) { // создать каталог если его нет
		err = os.MkdirAll(dir, os.ModePerm)
	}
	return err
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
