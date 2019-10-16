package main

import (
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/gotools/pkg/copydata"
	"github.com/fpawel/gotools/pkg/logfile"
	"github.com/fpawel/oxygen73/internal"
	"github.com/fpawel/oxygen73/internal/gui"
	"github.com/powerman/structlog"
	"os"
	"path/filepath"
)

func main() {
	log := structlog.New()

	structlog.DefaultLogger.
		SetPrefixKeys(
			structlog.KeyApp,
			structlog.KeyPID, structlog.KeyLevel, structlog.KeyUnit, structlog.KeyTime,
		).
		SetDefaultKeyvals(
			structlog.KeyApp, filepath.Base(os.Args[0]),
			structlog.KeySource, structlog.Auto,
		).
		SetSuffixKeys(
			structlog.KeyStack,
		).
		SetSuffixKeys(structlog.KeySource).
		SetKeysFormat(map[string]string{
			structlog.KeyTime:   " %[2]s",
			structlog.KeySource: " %6[2]s",
			structlog.KeyUnit:   " %6[2]s",
		})
	modbus.SetLogKeysFormat()
	guiWriter := copydata.NewWriter(gui.MsgWriteConsole, internal.WindowClass, internal.DelphiWindowClass)
	log.ErrIfFail(func() error {
		return logfile.Exec(guiWriter, "oxygen73.exe")
	})
}
