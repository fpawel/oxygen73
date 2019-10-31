package cfg

import (
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/oxygen73/internal/pkg/must"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Config struct {
	Main                  Hardware `yaml:"main"`
	Hum                   Hardware `yaml:"hum"`
	LogComm               bool     `yaml:"log_comm"`
	SaveMeasurementsCount int      `yaml:"save_measurements_count"`
}

type Hardware struct {
	Comport            string        `yaml:"comport"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"`
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"`
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`
	Pause              time.Duration `yaml:"pause"`
}

func (x Hardware) Comm() comm.Config {
	return comm.Config{
		TimeoutGetResponse: x.TimeoutGetResponse,
		TimeoutEndResponse: x.TimeoutEndResponse,
		MaxAttemptsRead:    x.MaxAttemptsRead,
		Pause:              x.Pause,
	}
}

func SetYaml(strYaml string) error {
	if err := yaml.Unmarshal([]byte(strYaml), &config); err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	comm.SetEnableLog(config.LogComm)
	mustWrite([]byte(strYaml))
	return nil
}

func GetYaml() string {
	mu.Lock()
	defer mu.Unlock()
	return string(must.MarshalYaml(&config))
}

func Set(v Config) {
	mu.Lock()
	defer mu.Unlock()
	data := must.MarshalYaml(&v)
	must.UnmarshalYaml(data, &config)
	comm.SetEnableLog(config.LogComm)
	mustWrite(data)
	return
}

func Get() (result Config) {
	mu.Lock()
	defer mu.Unlock()
	must.UnmarshalYaml(must.MarshalYaml(&config), &result)
	return
}

func mustWrite(b []byte) {
	must.WriteFile(filename(), b, 0666)
}

func filename() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "config.toml")
}

var (
	mu     sync.Mutex
	config = func() Config {
		x := Config{
			SaveMeasurementsCount: 20,
			LogComm:               false,
			Main: Hardware{
				TimeoutEndResponse: 50 * time.Millisecond,
				TimeoutGetResponse: 700 * time.Millisecond,
				MaxAttemptsRead:    3,
				Comport:            "COM1",
			},
			Hum: Hardware{
				TimeoutEndResponse: 50 * time.Millisecond,
				TimeoutGetResponse: 700 * time.Millisecond,
				MaxAttemptsRead:    3,
				Comport:            "COM2",
			},
		}
		filename := filename()

		mustWrite := func() {
			mustWrite(must.MarshalYaml(&x))
		}

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			mustWrite()
		}

		data, err := ioutil.ReadFile(filename)
		must.PanicIf(err)

		if err = yaml.Unmarshal(data, &x); err != nil {
			fmt.Println(err, "file:", filename)
			mustWrite()
		}

		comm.SetEnableLog(x.LogComm)
		return x
	}()
)
