//go:build linux

package main

import (
	"os/exec"
	"sort"
	"strings"
)

func GetSystemFonts() ([]FontInfo, error) {
	out, err := exec.Command("fc-list", "--format=%{family}\\n").Output()
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var fonts []FontInfo

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if idx := strings.Index(line, ","); idx != -1 {
			line = line[:idx]
		}
		parts := strings.Split(line, "&")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" || seen[part] {
				continue
			}
			seen[part] = true
			fonts = append(fonts, FontInfo{
				Family: part,
				Weight: 400,
				Italic: false,
			})
		}
	}

	sort.Slice(fonts, func(i, j int) bool {
		return fonts[i].Family < fonts[j].Family
	})
	return fonts, nil
}
