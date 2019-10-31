package must

import (
	"database/sql"
	"github.com/fpawel/oxygen73/internal/pkg"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"syscall"
)

// AbortIf should point to FatalIf or PanicIf or similar user-provided
// function which will interrupt execution in case it's param is not nil.
var AbortIf = PanicIf

// Alternative names to consider:
//  must.OK()
//  must.BeNil()
//  must.OrPanic()
//  must.OrAbort()
//  must.OrDie()

// NoErr is just a synonym for AbortIf.
func NoErr(err error) {
	AbortIf(err)
}

// PanicIf will call panic(err) in case given err is not nil.
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

// Decoder is an interface compatible with json.Decoder, gob.Decoder,
// xml.Decoder, …
type Decoder interface {
	Decode(v interface{}) error
}

// Encoder is an interface compatible with json.Encoder, gob.Encoder,
// xml.Encoder, …
type Encoder interface {
	Encode(v interface{}) error
}

// UnmarshalYAML is a wrapper for json.Unmarshal.
func UnmarshalYaml(data []byte, v interface{}) {
	err := yaml.Unmarshal(data, v)
	PanicIf(err)
}

// MarshalYaml is a wrapper for toml.Marshal.
func MarshalYaml(v interface{}) []byte {
	data, err := yaml.Marshal(v)
	PanicIf(err)
	return data
}

// WriteFile is a wrapper for ioutil.WriteFile.
func WriteFile(name string, buf []byte, perm os.FileMode) {
	err := ioutil.WriteFile(name, buf, perm)
	AbortIf(err)
}

func UTF16PtrFromString(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return p
}

//func EnsureDir(dir string) {
//	AbortIf(pkg.EnsureDir(dir))
//}

func OpenSqliteDBx(fileName string) *sqlx.DB {
	return sqlx.NewDb(OpenSqliteDB(fileName), "sqlite3")
}

func OpenSqliteDB(fileName string) *sql.DB {
	conn, err := pkg.OpenSqliteDB(fileName)
	NoErr(err)
	return conn
}
