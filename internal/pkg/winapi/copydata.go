package winapi

import (
	"github.com/fpawel/oxygen73/internal/pkg"
	"github.com/fpawel/oxygen73/internal/pkg/must"
	"github.com/lxn/win"
	"reflect"
	"unsafe"
)

type CopyData struct {
	DwData uintptr
	CbData uint32
	LpData uintptr
}

func CopyDataSendMessage(hWndSource, hWndTarget win.HWND, wParam uintptr, b []byte) uintptr {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	cd := CopyData{
		CbData: uint32(header.Len),
		LpData: header.Data,
		DwData: uintptr(hWndSource),
	}
	return win.SendMessage(hWndTarget, win.WM_COPYDATA, wParam, uintptr(unsafe.Pointer(&cd)))
}

func CopyDataSendString(hWndSource, hWndTarget win.HWND, msg uintptr, s string) uintptr {
	return CopyDataSendMessage(hWndSource, hWndTarget, msg, pkg.UTF16FromString(s))
}

func CopyDataSendJson(hWndSource, hWndTarget win.HWND, msg uintptr, param interface{}) uintptr {
	return CopyDataSendString(hWndSource, hWndTarget, msg, string(must.MarshalJSON(param)))
}

//func getCopyData(ptr unsafe.Pointer) (uintptr, []byte) {
//	cd := (*CopyData)(ptr)
//	p := PtrSliceFrom(unsafe.Pointer(cd.LpData), int(cd.CbData))
//	return cd.DwData, *(*[]byte)(p)
//}
//func ptrSliceFrom(p unsafe.Pointer, s int) unsafe.Pointer {
//	return unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(p), Len: s, Cap: s})
//}
