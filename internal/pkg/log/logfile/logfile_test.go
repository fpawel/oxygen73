package logfile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestListDays(t *testing.T) {
	logDir = filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "fpawel", "oxygen73", "build", "logs")
	days := ListDays()
	for _, t := range days {
		fmt.Println(t)
	}
	for _, x := range Read(days[len(days)-1]) {
		fmt.Println(x.Time, x.Line)
	}
}
