package main

import (
	"github.com/fpawel/oxygen73/internal/gui"
	"github.com/fpawel/oxygen73/internal/pkg/template/guintf"
	"os"
	"path/filepath"
	"reflect"
)

func main() {
	inFilename := filepath.Join(os.Getenv("GOPATH"),
		"src", "github.com", "fpawel", "oxygen73", "internal", "gui", "gui.go")
	inType := reflect.TypeOf((*gui.W)(nil)).Elem()
	outFilename := filepath.Join(os.Getenv("DELPHIPATH"),
		"src", "github.com", "fpawel", "oxygen73gui", "api", "api.notify.pas")
	guintf.Config{
		InFilename:  inFilename,
		OutFilename: outFilename,
		InType:      inType,
	}.MustWriteOutFile()
}
