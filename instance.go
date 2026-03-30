package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ensureSingleInstance() (bool, func()) {
	lockPath := filepath.Join(os.Getenv("APPDATA"), "gamzamemo", "app.lock")
	os.MkdirAll(filepath.Dir(lockPath), 0755)

	// 락 파일이 이미 있으면 PID 확인
	if data, err := os.ReadFile(lockPath); err == nil {
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err == nil && isProcessRunning(pid) {
			// 실제로 살아있는 프로세스 → 이미 실행 중
			return false, nil
		}
		// 죽은 프로세스의 락 파일 → 삭제하고 계속
		os.Remove(lockPath)
	}

	// 락 파일에 현재 PID 저장
	pid := os.Getpid()
	err := os.WriteFile(lockPath, []byte(fmt.Sprintf("%d", pid)), 0600)
	if err != nil {
		return false, nil
	}

	cleanup := func() {
		os.Remove(lockPath)
	}
	return true, cleanup
}
