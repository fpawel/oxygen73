package gui

import (
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
	MsgMeasurement
)

func WriteConsole(str string) bool {
	return w.SendString(MsgWriteConsole, str)
}

func Status(m StatusMessage) bool {
	return w.SendJson(MsgStatus, m)
}

func Measurement(m data.Measurement) bool {
	return w.SendJson(MsgMeasurement, m)
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
