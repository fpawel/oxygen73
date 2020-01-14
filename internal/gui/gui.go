package gui

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fpawel/gotools/pkg/copydata"
	"github.com/fpawel/oxygen73/internal"
	"github.com/fpawel/oxygen73/internal/data"
	"github.com/fpawel/oxygen73/internal/pkg/must"
	"github.com/fpawel/oxygen73/internal/pkg/winapi"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"strings"
	"time"
)

type Msg = uintptr

const (
	MsgWriteConsole Msg = iota
	MsgStatusComport
	MsgStatusComportHum
	MsgNewMeasurements
	MsgMeasurements
	MsgProductMeasurements
	MsgErrorOccurred
	MsgVacuumBegin
	MsgVacuumEnd
)

//func WriteConsole(str string) bool {
//	return w.SendString(MsgWriteConsole, str)
//}

func statusComport(m StatusMessage) bool {
	return w.SendJson(MsgStatusComport, m)
}
func statusComportHum(m StatusMessage) bool {
	return w.SendJson(MsgStatusComportHum, m)
}

func MsgBox(title, message string, style int) int {
	hWnd := win.FindWindow(must.UTF16PtrFromString(internal.DelphiWindowClass), nil)
	if hWnd == win.HWND_TOP {
		return 0
	}
	return int(win.MessageBox(
		hWnd,
		must.UTF16PtrFromString(strings.ReplaceAll(message, "\x00", "␀")),
		must.UTF16PtrFromString(strings.ReplaceAll(title, "\x00", "␀")),
		uint32(style)))
}

func VacuumBegin() bool {
	return w.SendString(MsgVacuumBegin, "")
}

func VacuumEnd() bool {
	return w.SendString(MsgVacuumEnd, "")
}

func Measurements(bucketID int64, ms []data.Measurement) bool {
	log.Debug(fmt.Sprintf("showing bucket %d: %d measurements", bucketID, len(ms)))
	t := time.Now()

	for n := 0; n < len(ms); {
		p := ms[n:]
		offset := len(p)
		if offset > 10000 {
			offset = 10000
		}
		p = p[:offset]
		n += offset

		buf := new(bytes.Buffer)
		writeBinary(buf, bucketID)
		writeBinary(buf, int64(len(p)))
		for _, m := range p {
			writeMeasurement(buf, m)
		}
		if !w.SendMessage(MsgMeasurements, buf.Bytes()) {
			return false
		}
	}

	buf := new(bytes.Buffer)
	writeBinary(buf, bucketID)
	writeBinary(buf, int64(0))
	if !w.SendMessage(MsgMeasurements, buf.Bytes()) {
		return false
	}

	log.Debug(fmt.Sprintf("bucket %d: %d measurements: %v", bucketID, len(ms), time.Since(t)))
	return true
}

func ErrorOccurred(err error) bool {
	return w.SendString(MsgErrorOccurred, err.Error())
}

func ProductMeasurements(bucketID int64, ms []data.Measurement1) bool {
	log.Debug(fmt.Sprintf("bucket %d: %d measurements", bucketID, len(ms)))
	buf := new(bytes.Buffer)
	writeBinary(buf, bucketID)
	writeBinary(buf, int64(len(ms)))
	for _, m := range ms {
		writeBinary(buf, m.StoredAt.UnixNano()/1000000) // количество миллисекунд метки времени
		writeBinary(buf, m.Temperature)
		writeBinary(buf, m.Pressure)
		writeBinary(buf, m.Humidity)
		writeBinary(buf, m.Value)
	}
	return w.SendMessage(MsgProductMeasurements, buf.Bytes())
}

func NewMeasurements(bucketID int64, ms []data.Measurement) bool {
	buf := new(bytes.Buffer)
	writeBinary(buf, bucketID)
	writeBinary(buf, int64(len(ms)))
	for _, m := range ms {
		writeMeasurement(buf, m)
	}
	return w.SendMessage(MsgNewMeasurements, buf.Bytes())
}

func StatusComportOk(text string) bool {
	return statusComport(StatusMessage{Ok: true, Text: text})
}

func StatusComportErr(err error) bool {
	return statusComport(StatusMessage{Ok: false, Text: err.Error()})
}

func StatusComportHumOk(text string) bool {
	return statusComportHum(StatusMessage{Ok: true, Text: text})
}

func StatusComportHumErr(err error) bool {
	return statusComportHum(StatusMessage{Ok: false, Text: err.Error()})
}

type StatusMessage struct {
	Ok   bool
	Text string
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
