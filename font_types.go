package main

type FontInfo struct {
	Family  string `json:"family"`
	Weight  int    `json:"weight"`
	Italic  bool   `json:"italic"`
	Stretch int    `json:"stretch"`
	Name    string `json:"name"`
}
