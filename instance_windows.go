package main

import "golang.org/x/sys/windows"

func isProcessRunning(pid int) bool {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(handle)

	var code uint32
	err = windows.GetExitCodeProcess(handle, &code)
	return err == nil && code == 259 // 259 = STILL_ACTIVE
}
