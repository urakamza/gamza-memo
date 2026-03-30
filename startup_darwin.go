package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const appName = "GamzaMemo"

func plistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", "com.gamzamemo.app.plist")
}

func RegisterStartup() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.gamzamemo.app</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>`, exePath)

	return os.WriteFile(plistPath(), []byte(plist), 0644)
}

func UnregisterStartup() error {
	return os.Remove(plistPath())
}

func IsStartupRegistered() bool {
	_, err := os.Stat(plistPath())
	return err == nil
}
