//go:build darwin

package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func GetSystemFonts() ([]FontInfo, error) {
	home, _ := os.UserHomeDir()
	dirs := []string{
		"/System/Library/Fonts",
		"/Library/Fonts",
		filepath.Join(home, "Library/Fonts"),
	}

	seen := map[string]bool{}
	var fonts []FontInfo

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			ext := strings.ToLower(filepath.Ext(name))
			if ext != ".ttf" && ext != ".otf" && ext != ".ttc" {
				continue
			}
			name = strings.TrimSuffix(name, filepath.Ext(name))
			parts := strings.Split(name, "&")
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
	}

	sort.Slice(fonts, func(i, j int) bool {
		return fonts[i].Family < fonts[j].Family
	})
	return fonts, nil
}
