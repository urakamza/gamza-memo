package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const appName = "GamzaMemo"

func desktopPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "autostart", "gamzamemo.desktop")
}

func RegisterStartup() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	desktop := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=GamzaMemo
Exec=%s
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
`, exePath)

	dir := filepath.Dir(desktopPath())
	os.MkdirAll(dir, 0755)
	return os.WriteFile(desktopPath(), []byte(desktop), 0644)
}

func UnregisterStartup() error {
	return os.Remove(desktopPath())
}

func IsStartupRegistered() bool {
	_, err := os.Stat(desktopPath())
	return err == nil
}
