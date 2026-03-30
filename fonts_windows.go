//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unicode/utf16"
	"unsafe"

	_ "embed"
)

//go:embed fonts_dwrite.dll
var dwriteDLL []byte

type dwriteFontWeight struct {
	Weight  int    `json:"weight"`
	Italic  bool   `json:"italic"`
	Stretch int    `json:"stretch"`
	Name    string `json:"name"`
}

type dwriteFontFamily struct {
	Family string             `json:"family"`
	Fonts  []dwriteFontWeight `json:"fonts"`
}

func loadDWriteDLL() (*syscall.DLL, error) {
	dir := filepath.Join(os.Getenv("APPDATA"), "gamzamemo")
	os.MkdirAll(dir, 0755)
	dllPath := filepath.Join(dir, "fonts_dwrite.dll")

	// 크기가 다르면 새 버전으로 교체
	needWrite := true
	if info, err := os.Stat(dllPath); err == nil {
		if info.Size() == int64(len(dwriteDLL)) {
			needWrite = false
		}
	}

	if needWrite {
		if err := os.WriteFile(dllPath, dwriteDLL, 0644); err != nil {
			return nil, fmt.Errorf("DLL 추출 실패: %w", err)
		}
	}

	return syscall.LoadDLL(dllPath)
}

func utf16PtrToString(buf []uint16) string {
	n := 0
	for n < len(buf) && buf[n] != 0 {
		n++
	}
	return string(utf16.Decode(buf[:n]))
}

func GetSystemFonts() ([]FontInfo, error) {
	dll, err := loadDWriteDLL()
	if err != nil {
		return nil, err
	}
	defer dll.Release()

	proc, err := dll.FindProc("GetFontListJSON")
	if err != nil {
		return nil, err
	}

	bufSize := 1024 * 1024 * 8 // 8MB
	buf := make([]uint16, bufSize)

	ret, _, _ := proc.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(bufSize),
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetFontListJSON 실패")
	}

	jsonStr := utf16PtrToString(buf)

	var families []dwriteFontFamily
	if err := json.Unmarshal([]byte(jsonStr), &families); err != nil {
		return nil, err
	}

	var result []FontInfo
	for _, f := range families {
		for _, fw := range f.Fonts {
			result = append(result, FontInfo{
				Family:  f.Family,
				Weight:  fw.Weight,
				Italic:  fw.Italic,
				Stretch: fw.Stretch,
				Name:    fw.Name,
			})
		}
	}

	return result, nil
}
