package main

type Settings struct {
	StartupEnabled bool    `json:"startupEnabled"`
	FontFamily     string  `json:"fontFamily"`
	FontSize       int     `json:"fontSize"`
	FontWeight     int     `json:"fontWeight"`
	FontItalic     bool    `json:"fontItalic"`
	FontStretch    int     `json:"fontStretch"`
	LineHeight     int     `json:"lineHeight"`
	LetterSpacing  float64 `json:"letterSpacing"`
	ImgWidth       int     `json:"imgWidth"`
	MainWinX       int     `json:"mainWinX"`
	MainWinY       int     `json:"mainWinY"`
	MainWinWidth   int     `json:"mainWinWidth"`
	MainWinHeight  int     `json:"mainWinHeight"`
	Theme          string  `json:"theme"`
	ConfirmDelete  bool    `json:"confirmDelete"`
}

func defaultSettings() Settings {
	return Settings{
		StartupEnabled: false,
		FontFamily:     "'Malgun Gothic', 'Apple SD Gothic Neo', 'Noto Sans KR', sans-serif",
		FontSize:       14,
		FontWeight:     400,
		FontItalic:     false,
		FontStretch:    5,
		LineHeight:     160,
		LetterSpacing:  0,
		ImgWidth:       0,
		MainWinX:       100,
		MainWinY:       100,
		MainWinWidth:   300,
		MainWinHeight:  400,
		Theme:          "system",
		ConfirmDelete:  true,
	}
}
