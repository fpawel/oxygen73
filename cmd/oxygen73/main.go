package main

import (
	"flag"
	"fmt"
	"github.com/fpawel/oxygen73/internal"
	"github.com/fpawel/oxygen73/internal/app"
	"github.com/fpawel/oxygen73/internal/pkg"
	"github.com/powerman/structlog"
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
	pkg.InitLog()
	structlog.DefaultLogger.SetLogLevel(structlog.ParseLevel(*logLevel))
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
