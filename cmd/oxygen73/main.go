package main

import (
	"flag"
	"fmt"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/gotools/pkg/ccolor"
	"github.com/fpawel/gotools/pkg/rungo"
	"github.com/fpawel/oxygen73/internal"
	"github.com/fpawel/oxygen73/internal/app"
	"github.com/fpawel/oxygen73/internal/guiclient"
	"github.com/powerman/structlog"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {

	defaultLogLevelStr := os.Getenv("OXYGEN73_LOG_LEVEL")
	if len(strings.TrimSpace(defaultLogLevelStr)) == 0 {
		defaultLogLevelStr = "info"
	}

	defaultDevMode := parseBool(os.Getenv("OXYGEN73_DEVMODE"))

	logLevel := flag.String("log.level", defaultLogLevelStr, "log `level` (debug|info|warn|err)")
	devMode := flag.Bool("dev", defaultDevMode,
		fmt.Sprintf("development mode on(true|false), default in OXYGEN73_DEVMODE env var: %t", defaultDevMode))
	flag.Parse()

	internal.DevMode = *devMode

	logFileOutput := rungo.NewLogFileOutput()
	defer structlog.DefaultLogger.ErrIfFail(logFileOutput.Close)

	output := io.MultiWriter(ccolor.NewWriter(os.Stderr), logFileOutput, guiclient.WriterNotifyConsole())

	log.SetOutput(output)
	log.SetPrefix(fmt.Sprintf("[%d] inf    ", os.Getpid()))

	structlog.DefaultLogger.
		//SetLogFormat(structlog.JSON).
		//SetTimeFormat(time.RFC3339Nano).
		//SetTimeValFormat(time.RFC3339Nano).
		// Wrong log.level is not fatal, it will be reported and set to "debug".
		SetOutput(output).
		SetLogLevel(structlog.ParseLevel(*logLevel)).
		SetPrefixKeys(
			//structlog.KeyApp,
			structlog.KeyPID, structlog.KeyLevel, structlog.KeyUnit, structlog.KeyTime,
		).
		SetDefaultKeyvals(
			//structlog.KeyApp, filepath.Base(os.Args[0]),
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
	app.Main()
}

func parseBool(s string) bool {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return v
}

// todo: основной воркер, выполняющий опрос и запись в бд
// todo: GUI: логгирование

// todo: сервер GUI + клиент backend
// todo: 	- оповещение: новые измерения
// todo: 	- оповещение: сообщения консоли GUI
// todo: 	- оповещение: сообщения консоли GUI

// todo: настройки приложения в GUI

// todo: конфигурация toml в GUI

// todo: удаление букетов
