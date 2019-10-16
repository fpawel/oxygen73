package main

import (
	"flag"
	"github.com/fpawel/oxygen73/internal/app"
	"github.com/fpawel/oxygen73/internal/pkg"
	"github.com/powerman/structlog"
	"os"
	"strings"
)

func main() {

	defaultLogLevelStr := os.Getenv("OXYGEN73_LOG_LEVEL")
	if len(strings.TrimSpace(defaultLogLevelStr)) == 0 {
		defaultLogLevelStr = "info"
	}

	logLevel := flag.String("log.level", defaultLogLevelStr, "log `level` (debug|info|warn|err)")
	flag.Parse()
	pkg.InitLog()
	structlog.DefaultLogger.SetLogLevel(structlog.ParseLevel(*logLevel))
	app.Main()
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
