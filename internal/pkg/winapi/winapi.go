package winapi

import (
	"github.com/fpawel/oxygen73/internal/pkg/must"
	"github.com/lxn/win"
	"log"
	"syscall"
	"unsafe"
)

var (
	libUser32     = mustLoadLibrary("user32.dll")
	isWindow      = mustGetProcAddress(libUser32, "IsWindow")
	getClassNameW = mustGetProcAddress(libUser32, "GetClassNameW")
)

func FindOrCreateNewWindowWithClassName(windowClassName string) win.HWND {
	hWnd := FindWindow(windowClassName)
	if !IsWindow(hWnd) {
		hWnd = NewWindowWithClassName(windowClassName, win.DefWindowProc)
	}
	if !IsWindow(hWnd) {
		panic(windowClassName)
	}
	return hWnd
}

func IsWindow(hWnd win.HWND) bool {
	ret, _, _ := syscall.Syscall(isWindow, 1,
		uintptr(hWnd),
		0,
		0)

	return ret != 0
}

func FindWindow(className string) win.HWND {
	ptrClassName := must.UTF16PtrFromString(className)
	return win.FindWindow(ptrClassName, nil)
}

type WindowProcedure = func(hWnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr

func NewWindowWithClassName(windowClassName string, windowProcedure WindowProcedure) win.HWND {

	wndProc := func(hWnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
		switch msg {
		case win.WM_DESTROY:
			win.PostQuitMessage(0)
		default:
			return windowProcedure(hWnd, msg, wParam, lParam)
		}
		return 0
	}

	mustRegisterWindowClassWithWndProcPtr(
		windowClassName, syscall.NewCallback(wndProc))

	return win.CreateWindowEx(
		0,
		must.UTF16PtrFromString(windowClassName),
		nil,
		0,
		0,
		0,
		0,
		0,
		win.HWND_TOP,
		0,
		win.GetModuleHandle(nil),
		nil)
}

func mustRegisterWindowClassWithWndProcPtr(className string, wndProcPtr uintptr) {

	hInst := win.GetModuleHandle(nil)
	if hInst == 0 {
		panic("GetModuleHandle")
	}

	hIcon := win.LoadIcon(hInst, win.MAKEINTRESOURCE(7)) // rsrc uses 7 for app icon
	if hIcon == 0 {
		hIcon = win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_APPLICATION))
	}
	if hIcon == 0 {
		panic("LoadIcon")
	}

	hCursor := win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ARROW))
	if hCursor == 0 {
		panic("LoadCursor")
	}

	var wc win.WNDCLASSEX
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = wndProcPtr
	wc.HInstance = hInst
	wc.HIcon = hIcon
	wc.HCursor = hCursor
	wc.HbrBackground = win.COLOR_BTNFACE + 1
	wc.LpszClassName = syscall.StringToUTF16Ptr(className)
	wc.Style = 0

	if atom := win.RegisterClassEx(&wc); atom == 0 {
		panic("RegisterClassEx")
	}

}

func mustGetProcAddress(lib uintptr, name string) uintptr {
	addr, err := syscall.GetProcAddress(syscall.Handle(lib), name)
	if err != nil {
		log.Panicln("get procedure address:", name, ":", err)
	}

	return uintptr(addr)
}

func mustLoadLibrary(name string) uintptr {
	lib, err := syscall.LoadLibrary(name)
	if err != nil {
		log.Panicln("load library:", name, ":", err)
	}
	return uintptr(lib)
}

func mustLoadDLL(name string) *syscall.DLL {
	dll, err := syscall.LoadDLL("Advapi32.dll")
	if err != nil {
		log.Panicln("load dll:", name, ":", err)
	}
	return dll
}

func mustFindProc(dll *syscall.DLL, name string) *syscall.Proc {
	proc, err := dll.FindProc(name)
	if err != nil {
		log.Panicln("find procedure address:", name, ":", err)
	}
	return proc
}
