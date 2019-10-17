package gui

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fpawel/gotools/pkg/copydata"
	"github.com/fpawel/oxygen73/internal"
	"github.com/fpawel/oxygen73/internal/data"
	"github.com/fpawel/oxygen73/internal/pkg/winapi"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"strings"
)

type Msg = uintptr

const (
	MsgWriteConsole Msg = iota
	MsgStatus
	MsgNewMeasurements
	MsgMeasurements
)

//func WriteConsole(str string) bool {
//	return w.SendString(MsgWriteConsole, str)
//}

func Status(m StatusMessage) bool {
	return w.SendJson(MsgStatus, m)
}

func Measurements(ms []data.Measurement) bool {
	buf := new(bytes.Buffer)
	writeBinary(buf, int64(len(ms)))
	for _, m := range ms {
		writeMeasurement(buf, m)
	}
	return w.SendMessage(MsgMeasurements, buf.Bytes())
}

func NewMeasurements(ms []data.Measurement) bool {
	buf := new(bytes.Buffer)
	writeBinary(buf, int64(len(ms)))
	for _, m := range ms {
		writeMeasurement(buf, m)
	}
	return w.SendMessage(MsgNewMeasurements, buf.Bytes())
}

func StatusOk(text string) bool {
	return Status(StatusMessage{Ok: true, Text: text})
}

func StatusErr(err error) bool {
	return Status(StatusMessage{Ok: false, Text: cutErrStr(err), Detail: err.Error()})
}

type StatusMessage struct {
	Ok           bool
	Text, Detail string
}

func cutErrStr(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if strings.Contains(s, ":") {
		return strings.Split(s, ":")[0]
	}
	return s
}

// Handle выполняет бесконечный цикл с обработкой оконных сообщений.
// Необходим для работы механизма отправки сообщений WM_COPYDATA
func Handle() {

	// инициализация окна окно связи с GUI для отправки сообщений WM_COPYDATA
	winapi.NewWindowWithClassName(internal.WindowClass, win.DefWindowProc)

	for {
		var msg win.MSG
		if win.GetMessage(&msg, 0, 0, 0) == 0 {
			log.Info("выход из цикла оконных сообщений")
			return
		}
		log.Debug(fmt.Sprintf("%+v", msg))
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
}

var (
	log = structlog.New()
	w   = copydata.WndClass{
		Src:  internal.WindowClass,
		Dest: internal.DelphiWindowClass,
	}
)

func writeMeasurement(buf *bytes.Buffer, m data.Measurement) {
	writeBinary(buf, m.StoredAt.UnixNano()/1000000) // количество миллисекунд метки времени
	writeBinary(buf, m.Temperature)
	writeBinary(buf, m.Pressure)
	writeBinary(buf, m.Humidity)
	for i := 0; i < 50; i++ {
		writeBinary(buf, m.Places[i])
	}
}

func writeBinary(buf *bytes.Buffer, data interface{}) {
	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		log.Fatal(err)
	}
}
