package cfg

import (
	"github.com/fpawel/comm"
	"github.com/fpawel/oxygen73/internal/pkg/must"
	"github.com/pelletier/go-toml"
	"github.com/powerman/structlog"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	Main                  Hardware `toml:"main" comment:"параметры стенда"`
	Hum                   Hardware `toml:"hum" comment:"параметры дадчика влажности"`
	LogComport            bool     `toml:"log_comport" comment:"логгирование посылок COM порта"`
	SaveMeasurementsCount int      `toml:"save_measurements_count" comment:"количество сохраняемых измерений"`
}

type Hardware struct {
	Comm    comm.Config `toml:"comm" comment:"параметры приёмопередачи прибора"`
	Comport string      `toml:"comport" comment:"имя СОМ порта проибора"`
}

func Open(log *structlog.Logger) {
	defer func() {
		comm.SetEnableLog(config.LogComport)
	}()
	if _, err := os.Stat(fileName()); os.IsNotExist(err) {
		save()
		return
	}
	data, err := ioutil.ReadFile(fileName())
	must.AbortIf(err)
	if err = toml.Unmarshal(data, &config); err != nil {
		log.PrintErr(err, "file", fileName())
	}
}

func SetToml(strToml string) error {
	mu.Lock()
	defer mu.Unlock()
	if err := toml.Unmarshal([]byte(strToml), &config); err != nil {
		return err
	}
	comm.SetEnableLog(config.LogComport)
	write([]byte(strToml))
	return nil
}

func Set(v Config) {
	mu.Lock()
	defer mu.Unlock()
	must.UnmarshalJSON(must.MarshalJSON(&v), &config)
	comm.SetEnableLog(config.LogComport)
	save()
	return
}

func Get() (result Config) {
	mu.Lock()
	defer mu.Unlock()
	must.UnmarshalJSON(must.MarshalJSON(&config), &result)
	return
}

func fileName() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "config.toml")
}
func save() {
	write(must.MarshalToml(config))
}
func write(data []byte) {
	must.WriteFile(fileName(), data, 0666)
}

var (
	mu     sync.Mutex
	config = Config{
		SaveMeasurementsCount: 20,
		LogComport:            false,
		Main: Hardware{
			Comm: comm.Config{
				ReadByteTimeoutMillis: 50,
				ReadTimeoutMillis:     700,
				MaxAttemptsRead:       3,
			},
			Comport: "COM1",
		},
		Hum: Hardware{
			Comm: comm.Config{
				ReadByteTimeoutMillis: 50,
				ReadTimeoutMillis:     700,
				MaxAttemptsRead:       3,
			},
			Comport: "COM2",
		},
	}
)
