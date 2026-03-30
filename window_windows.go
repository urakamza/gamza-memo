//go:build windows

package main

import (
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	getWindowLong = user32.NewProc("GetWindowLongW")
	setWindowLong = user32.NewProc("SetWindowLongW")
	showWindow    = user32.NewProc("ShowWindow")
	flashWindowEx = user32.NewProc("FlashWindowEx")
)

const (
	WS_EX_NOACTIVATE  = uintptr(0x08000000)
	SW_SHOWNOACTIVATE = 4
	FLASHW_STOP       = 0
)

type FLASHWINFO struct {
	CbSize    uint32
	Hwnd      uintptr
	DwFlags   uint32
	UCount    uint32
	DwTimeout uint32
}

func stopFlash(win *application.WebviewWindow) {
	hwnd := uintptr(win.NativeWindow())
	info := FLASHWINFO{
		DwFlags: uint32(FLASHW_STOP),
		Hwnd:    hwnd,
		UCount:  0,
	}
	info.CbSize = uint32(unsafe.Sizeof(info))
	flashWindowEx.Call(uintptr(unsafe.Pointer(&info)))
}

func showNoActivate(win *application.WebviewWindow) {
	hwnd := uintptr(win.NativeWindow())
	showWindow.Call(hwnd, SW_SHOWNOACTIVATE)
}
