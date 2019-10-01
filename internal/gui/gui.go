package gui

import (
	"fmt"
	"github.com/fpawel/oxygen73/internal"
	"github.com/fpawel/oxygen73/internal/data"
	"github.com/fpawel/oxygen73/internal/pkg/winapi"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"runtime"
	"strings"
)

type Msg = uintptr

const (
	MsgWriteConsole Msg = iota
	MsgStatus
	MsgMeasurement
)

type W struct{}

func (x W) WriteConsole(str string) uintptr {
	return winapi.CopyDataSendString(HWndSource(), hWndTarget(), MsgWriteConsole, str)
}

func (x W) Status(m StatusMessage) uintptr {
	return winapi.CopyDataSendJson(HWndSource(), hWndTarget(), MsgStatus, m)
}

func (x W) Measurement(m data.Measurement) uintptr {
	return winapi.CopyDataSendJson(HWndSource(), hWndTarget(), MsgMeasurement, m)
}

func StatusOk(text string) uintptr {
	return W{}.Status(StatusMessage{Ok: true, Text: text})
}

func StatusErr(err error) uintptr {
	return W{}.Status(StatusMessage{Ok: false, Text: cutErrStr(err), Detail: err.Error()})
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

func HWndSource() win.HWND {
	return winapi.FindWindow(internal.WindowClass)
}

func hWndTarget() win.HWND {
	return winapi.FindWindow(internal.DelphiWindowClass)
}

func Handle() {

	// цикл оконных сообщений. необходим для работы механизма отправки сообщений WM_COPYDATA
	// должен работать в том же потоке ОС, в котором был создан объект окна
	runtime.LockOSThread()

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
)
