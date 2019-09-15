package main

import (
	"fmt"
	"github.com/fpawel/gohelp/must"
	"github.com/powerman/structlog"
	"gopkg.in/ini.v1"
	"net"
	"os"
	"path/filepath"
)

func main() {

	influxDir := filepath.Join(filepath.Dir(os.Args[0]), "influxdb")
	influxPth := func(s ...string) string {
		return filepath.Join(append([]string{influxDir}, s...)...)
	}
	mustInfluxDir := func(s ...string) string {
		dir := influxPth(s...)
		must.EnsureDir(dir)
		return dir
	}
	confFilePath := influxPth("influxdb.conf")
	conf, err := ini.Load(confFilePath)
	if err != nil {
		panic(fmt.Sprintf("Fail to read file %q: %v", confFilePath, err))
	}
	os.Setenv("INFLUXDB_DATA_DIR", mustInfluxDir("data"))

	conf.Section("meta").Key("dir").SetValue(mustInfluxDir("meta"))
	conf.Section("data").Key("dir").SetValue(mustInfluxDir("data"))
	conf.Section("data").Key("wal-dir").SetValue(mustInfluxDir("wal"))
	conf.Section("http").DeleteKey("enabled")

	influxPort := influxPort()
	if influxPort == defaultInfluxPort {
		conf.Section("http").DeleteKey("bind-address")
	} else {
		conf.Section("http").Key("bind-address").SetValue(fmt.Sprintf(":%d", influxPort))
	}
	if err := conf.SaveTo(confFilePath); err != nil {
		panic(fmt.Sprintf("Save config %q: %v", confFilePath, err))
	}
}

func init() {
	structlog.DefaultLogger.
		SetPrefixKeys(structlog.KeyApp, structlog.KeyPID, structlog.KeyLevel, structlog.KeyUnit).
		SetSuffixKeys(structlog.KeySource, structlog.KeyStack).
		SetDefaultKeyvals(
			structlog.KeyApp, filepath.Base(os.Args[0]),
			structlog.KeySource, structlog.Auto,
		).
		SetKeysFormat(map[string]string{
			structlog.KeySource: " %6[2]s",
			structlog.KeyUnit:   " %6[2]s",
		})
}
func influxPort() int {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", defaultInfluxPort))
	if err == nil {
		if err = ln.Close(); err != nil {
			panic(err)
		}
		return defaultInfluxPort
	}
	if ln, err = net.Listen("tcp", "127.0.0.1:0"); err != nil {
		panic(err)
	}
	influxPort := ln.Addr().(*net.TCPAddr).Port
	if err = ln.Close(); err != nil {
		panic(err)
	}
	return influxPort
}

const (
	defaultInfluxPort = 8086
)
