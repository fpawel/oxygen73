package cfg

import (
	"encoding/json"
	"github.com/fpawel/comm"
	"github.com/fpawel/gohelp/must"
	"github.com/powerman/structlog"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	Public   Public
	Internal Internal
}

type Public struct {
	ComportName           string      `toml:"comport_name" comment:"Имя СОМ порта"`
	LogComport            bool        `toml:"log_comport" comment:"логгирование посылок COM порта"`
	SaveMeasurementsCount int         `toml:"save_measurements_count" comment:"количество сохраняемых измерений"`
	Comm                  comm.Config `toml:"comm" comment:"параметры приёмо-передачи"`
}

type Internal struct {
	ActiveSeries []bool
}

func Open(log *structlog.Logger) {
	b, err := ioutil.ReadFile(fileName())
	if err == nil {
		err = json.Unmarshal(b, &config)
	}
	if err != nil {
		log.PrintErr(err, "file", fileName())
	}
	comm.SetEnableLog(config.Public.LogComport)

	// создать файл конфигурации, если его нет
	if _, err := os.Stat(fileName()); os.IsNotExist(err) {
		must.WriteFile(fileName(), must.MarshalIndentJSON(&config, "", "    "), 0666)
	}

}

func Setup(v Config) {
	mu.Lock()
	defer mu.Unlock()
	comm.SetEnableLog(v.Public.LogComport)
	must.UnmarshalJSON(must.MarshalJSON(&v), &config)
	must.WriteFile(fileName(), must.MarshalIndentJSON(&config, "", "    "), 0666)
	return
}

func Get() (result Config) {
	mu.Lock()
	defer mu.Unlock()
	must.UnmarshalJSON(must.MarshalJSON(&config), &result)
	return
}

func fileName() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "config.json")
}

var (
	mu     sync.Mutex
	config = Config{
		Public: Public{
			SaveMeasurementsCount: 20,
			ComportName:           "COM1",
			LogComport:            false,
			Comm: comm.Config{
				ReadByteTimeoutMillis: 50,
				ReadTimeoutMillis:     700,
				MaxAttemptsRead:       3,
			},
		},
	}
)
