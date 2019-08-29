package internal

import (
	"os"
	"path/filepath"
)

func DataDir() string {
	if os.Getenv("OXYGEN73_DEV_DB") == "true" {
		return filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "fpawel", "oxygen73", "build")
	}
	return filepath.Dir(os.Args[0])
}
