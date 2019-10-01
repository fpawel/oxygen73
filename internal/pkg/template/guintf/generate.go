package guintf

import (
	"github.com/fpawel/oxygen73/internal/pkg/must"
	"os"
	"path/filepath"
	"reflect"
)

type Config struct {
	InFilename, OutFilename string
	InType                  reflect.Type
}

func (x Config) MustWriteOutFile() {
	dir := filepath.Dir(x.OutFilename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		must.AbortIf(os.MkdirAll(dir, os.ModePerm))
	}
	file := must.Create(x.OutFilename)
	functions := parseFunctions(x.InFilename, x.InType)
	unit := newUnit(filepath.Base(x.OutFilename), functions)
	unit.writeunit(file)
	must.Close(file)
}
