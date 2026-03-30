package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

type NoteService struct {
	db           *sql.DB
	app          *application.App
	windows      map[string]*application.WebviewWindow
	mainWindow   *application.WebviewWindow
	updateTimers map[string]*time.Timer
	pendingNotes map[string]Note
	mu           sync.Mutex
	settings     Settings
}
type Note struct {
	Id          string `json:"id"`
	Content     string `json:"content"`
	Color       string `json:"color"`
	WinX        int    `json:"winX"`
	WinY        int    `json:"winY"`
	WinWidth    int    `json:"winWidth"`
	WinHeight   int    `json:"winHeight"`
	IsOpen      bool   `json:"isOpen"`
	AlwaysOnTop bool   `json:"alwaysOnTop"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}

func (s *NoteService) OpenMainWindow() {
	if s.mainWindow != nil {
		s.mainWindow.Focus()
		return
	}

	x := s.settings.MainWinX
	y := s.settings.MainWinY
	width := s.settings.MainWinWidth
	height := s.settings.MainWinHeight
	if width == 0 {
		width = 300
	}
	if height == 0 {
		height = 400
	}

	win := s.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "감자 메모",
		Width:     width,
		Height:    height,
		MinWidth:  250,
		MinHeight: 200,
		Frameless: true,
		X:         x,
		Y:         y,
		URL:       "/",
	})
	s.mainWindow = win

	win.OnWindowEvent(events.Common.WindowShow, func(e *application.WindowEvent) {
		win.SetPosition(x, y)
	})

	var moveTimer *time.Timer
	win.OnWindowEvent(events.Common.WindowDidMove, func(e *application.WindowEvent) {
		if moveTimer != nil {
			moveTimer.Stop()
		}
		moveTimer = time.AfterFunc(500*time.Millisecond, func() {
			bounds := win.Bounds()
			s.settings.MainWinX = bounds.X
			s.settings.MainWinY = bounds.Y
			s.saveSettings()
		})
	})

	var sizeTimer *time.Timer
	win.OnWindowEvent(events.Common.WindowDidResize, func(e *application.WindowEvent) {
		if sizeTimer != nil {
			sizeTimer.Stop()
		}
		sizeTimer = time.AfterFunc(500*time.Millisecond, func() {
			bounds := win.Bounds()
			s.settings.MainWinWidth = bounds.Width
			s.settings.MainWinHeight = bounds.Height
			s.saveSettings()
		})
	})

	win.OnWindowEvent(events.Common.WindowClosing, func(e *application.WindowEvent) {
		bounds := win.Bounds()
		s.settings.MainWinX = bounds.X
		s.settings.MainWinY = bounds.Y
		s.settings.MainWinWidth = bounds.Width
		s.settings.MainWinHeight = bounds.Height
		s.saveSettings()
		s.mainWindow = nil
	})
}

func NewNoteService() *NoteService {
	s := &NoteService{}
	s.updateTimers = make(map[string]*time.Timer)
	s.pendingNotes = make(map[string]Note)

	dbPath := filepath.Join(os.Getenv("APPDATA"), "gamzamemo", "notes.db")
	os.MkdirAll(filepath.Dir(dbPath), 0755)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	db.Exec(`PRAGMA journal_mode=WAL`)
	db.Exec((`PRAGMA busy_timeout=5000`))
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS notes (
		id TEXT PRIMARY KEY,
		content TEXT,
		color TEXT,
		win_x INTEGER,
		win_y INTEGER,
		win_width INTEGER,
		win_height INTEGER,
		is_open INTEGER,
		always_on_top INTEGER,
		created TEXT,
		updated TEXT
	)`)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS images (
		id TEXT PRIMARY KEY,
		note_id TEXT NOT NULL,
		filename TEXT,
		data BLOB NOT NULL,
		FOREIGN KEY (note_id) REFERENCES notes(id)
	)`)
	if err != nil {
		panic(err)
	}
	s.db = db
	s.windows = make(map[string]*application.WebviewWindow)
	s.loadSettings()
	return s
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (s *NoteService) GetNotes() ([]Note, error) {
	rows, err := s.db.Query(`SELECT id, content, color, win_x, win_y, win_width, win_height, is_open, always_on_top, created, updated FROM notes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		var isOpen, alwaysOnTop int
		err := rows.Scan(&n.Id, &n.Content, &n.Color, &n.WinX, &n.WinY, &n.WinWidth, &n.WinHeight, &isOpen, &alwaysOnTop, &n.Created, &n.Updated)
		if err != nil {
			return nil, err
		}
		n.IsOpen = isOpen == 1
		n.AlwaysOnTop = alwaysOnTop == 1
		notes = append(notes, n)
	}
	return notes, nil
}

func (s *NoteService) GetNoteById(id string) (Note, error) {
	var note Note
	var isOpen, alwaysOnTop int
	err := s.db.QueryRow(`SELECT id, content, color, win_x, win_y, win_width, win_height, is_open, always_on_top, created, updated FROM notes WHERE id=?`, id).Scan(
		&note.Id, &note.Content, &note.Color, &note.WinX, &note.WinY, &note.WinWidth, &note.WinHeight, &isOpen, &alwaysOnTop, &note.Created, &note.Updated,
	)
	if err != nil {
		return Note{}, err
	}
	note.IsOpen = isOpen == 1
	note.AlwaysOnTop = alwaysOnTop == 1
	return note, nil
}

func (s *NoteService) CreateNote() (Note, error) {
	offset := len(s.windows) * 30
	now := time.Now().Format(time.RFC3339)
	note := Note{
		Id:          fmt.Sprintf("%d", time.Now().UnixNano()),
		Content:     "",
		Color:       "#fff2c2",
		WinX:        200 + offset,
		WinY:        200 + offset,
		WinWidth:    200,
		WinHeight:   250,
		IsOpen:      true,
		AlwaysOnTop: false,
		Created:     now,
		Updated:     now,
	}
	_, err := s.db.Exec(`INSERT INTO notes (id, content, color, win_x, win_y, win_width, win_height, is_open, always_on_top, created, updated) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		note.Id, note.Content, note.Color, note.WinX, note.WinY, note.WinWidth, note.WinHeight,
		boolToInt(note.IsOpen), boolToInt(note.AlwaysOnTop), note.Created, note.Updated,
	)
	if err != nil {
		return Note{}, err
	}
	s.OpenNoteWindow(note.Id)
	s.emitNotesUpdated()
	return note, nil
}

func (s *NoteService) UpdateNote(note Note) (Note, error) {
	s.mu.Lock()
	s.pendingNotes[note.Id] = note
	if timer, exists := s.updateTimers[note.Id]; exists {
		timer.Stop()
	}
	s.updateTimers[note.Id] = time.AfterFunc(1500*time.Millisecond, func() {
		s.mu.Lock()
		n := s.pendingNotes[note.Id]
		s.mu.Unlock()

		now := time.Now().Format(time.RFC3339)
		s.db.Exec(`UPDATE notes SET content=?, color=?, win_x=?, win_y=?, win_width=?, win_height=?, is_open=?, always_on_top=?, updated=? WHERE id=?`,
			n.Content, n.Color, n.WinX, n.WinY, n.WinWidth, n.WinHeight,
			boolToInt(n.IsOpen), boolToInt(n.AlwaysOnTop), now, n.Id,
		)
		println("저장됨:", n.Id, time.Now().Format("15:04:05"))
		s.emitNotesUpdated()
	})
	s.mu.Unlock()
	return note, nil
}

func (s *NoteService) SaveNoteNow(note Note) error {
	s.mu.Lock()
	// 기존 디바운스 타이머 취소
	if timer, exists := s.updateTimers[note.Id]; exists {
		timer.Stop()
		delete(s.updateTimers, note.Id)
	}
	delete(s.pendingNotes, note.Id)
	s.mu.Unlock()

	now := time.Now().Format(time.RFC3339)
	_, err := s.db.Exec(`UPDATE notes SET content=?, color=?, win_x=?, win_y=?, win_width=?, win_height=?, is_open=?, always_on_top=?, updated=? WHERE id=?`,
		note.Content, note.Color, note.WinX, note.WinY, note.WinWidth, note.WinHeight,
		boolToInt(note.IsOpen), boolToInt(note.AlwaysOnTop), now, note.Id,
	)
	println("즉시 저장됨:", note.Id, time.Now().Format("15:04:05"))
	s.emitNotesUpdated()
	return err
}

func (s *NoteService) UpdateColor(note Note) (Note, error) {
	s.db.Exec(`UPDATE notes SET color=? WHERE id=?`,
		note.Color, note.Id,
	)
	println("색상 변경됨:", note.Id, time.Now().Format("15:04:05"))
	s.emitNotesUpdated()
	return note, nil
}

func (s *NoteService) UpdatePin(note Note) (Note, error) {
	s.db.Exec(`UPDATE notes SET always_on_top=? WHERE id=?`,
		boolToInt(note.AlwaysOnTop), note.Id,
	)
	println("핀 변경됨:", note.Id, time.Now().Format("15:04:05"))
	return note, nil
}

func (s *NoteService) CloseNote(id string) error {
	if win, exists := s.windows[id]; exists {
		win.Close()
	}
	return nil
}

func (s *NoteService) DeleteNote(id string) error {
	if win, exists := s.windows[id]; exists {
		win.Close()
		delete(s.windows, id)
	}
	// 이미지 DB에서 삭제
	_, err := s.db.Exec(`DELETE FROM images WHERE note_id=?`, id)
	if err != nil {
		return err
	}

	// 메모 삭제
	_, err = s.db.Exec(`DELETE FROM notes WHERE id=?`, id)
	if err != nil {
		return err
	}
	println("삭제됨:", id, time.Now().Format("15:04:05"))
	s.emitNotesUpdated()
	return nil
}

func (s *NoteService) OpenNoteWindow(id string) {
	if win, exists := s.windows[id]; exists {
		win.Focus()
		return
	}
	note, err := s.GetNoteById(id)
	if err != nil {
		return
	}
	width := note.WinWidth
	if width == 0 {
		width = 200
	}
	height := note.WinHeight
	if height == 0 {
		height = 250
	}
	x := note.WinX
	if x == 0 {
		x = 100
	}
	y := note.WinY
	if y == 0 {
		y = 100
	}
	win := s.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "메모",
		Width:          width,
		Height:         height,
		MinWidth:       200,
		MinHeight:      150,
		X:              x,
		Y:              y,
		URL:            "/?noteId=" + id,
		AlwaysOnTop:    note.AlwaysOnTop,
		EnableFileDrop: true,
		Frameless:      true,
		Hidden:         true,
	})
	s.windows[id] = win

	s.db.Exec(`UPDATE notes SET is_open=1 WHERE id=?`, id)
	s.emitNotesUpdated()

	isReady := false
	win.OnWindowEvent(events.Common.WindowShow, func(e *application.WindowEvent) {
		win.SetPosition(x, y)
		stopFlash(win)
		time.AfterFunc(500*time.Millisecond, func() {
			isReady = true
		})
	})

	var moveTimer *time.Timer
	win.OnWindowEvent(events.Common.WindowDidMove, func(e *application.WindowEvent) {
		if !isReady {
			return
		}
		if moveTimer != nil {
			moveTimer.Stop()
		}
		moveTimer = time.AfterFunc(500*time.Millisecond, func() {
			bounds := win.Bounds()
			s.db.Exec(`UPDATE notes SET win_x=?, win_y=? WHERE id=?`, bounds.X, bounds.Y, id)
			println("이동:", bounds.X, bounds.Y, id, time.Now().Format("15:04:05"))
		})
	})

	var sizeTimer *time.Timer
	win.OnWindowEvent(events.Common.WindowDidResize, func(e *application.WindowEvent) {
		if !isReady {
			return
		}
		if sizeTimer != nil {
			sizeTimer.Stop()
		}
		sizeTimer = time.AfterFunc(500*time.Millisecond, func() {
			bounds := win.Bounds()
			s.db.Exec(`UPDATE notes SET win_width=?, win_height=? WHERE id=?`, bounds.Width, bounds.Height, id)
			println("리사이즈:", bounds.Width, bounds.Height, id, time.Now().Format("15:04:05"))
		})
	})
	win.OnWindowEvent(events.Common.WindowClosing, func(e *application.WindowEvent) {
		bounds := win.Bounds()

		s.mu.Lock()
		if timer, exists := s.updateTimers[id]; exists {
			timer.Stop()
			delete(s.updateTimers, id)
		}
		pending, hasPending := s.pendingNotes[id]
		if hasPending {
			delete(s.pendingNotes, id)
		}
		s.mu.Unlock()

		if hasPending {
			pending.WinX = bounds.X
			pending.WinY = bounds.Y
			pending.WinWidth = bounds.Width
			pending.WinHeight = bounds.Height
			pending.IsOpen = false
			now := time.Now().Format(time.RFC3339)
			s.db.Exec(`UPDATE notes SET content=?, color=?, win_x=?, win_y=?, win_width=?, win_height=?, is_open=?, always_on_top=?, updated=? WHERE id=?`,
				pending.Content, pending.Color, pending.WinX, pending.WinY, pending.WinWidth, pending.WinHeight,
				boolToInt(pending.IsOpen), boolToInt(pending.AlwaysOnTop), now, id,
			)
		} else {
			s.db.Exec(`UPDATE notes SET win_x=?, win_y=?, win_width=?, win_height=?, is_open=0 WHERE id=?`,
				bounds.X, bounds.Y, bounds.Width, bounds.Height, id)
		}
		delete(s.windows, id)
		s.emitNotesUpdated()
		println("창 닫힘:", time.Now().Format("15:04:05"))
	})
	go func() {
		time.Sleep(500 * time.Millisecond)
		showNoActivate(win)
		win.Show()
	}()
}

func (s *NoteService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	notes, err := s.GetNotes()
	if err != nil {
		return err
	}
	go func() {
		time.Sleep(500 * time.Millisecond)
		openCount := 0
		for _, note := range notes {
			if note.IsOpen {
				s.OpenNoteWindow(note.Id)
				openCount++
			}
		}
		if openCount == 0 {
			s.OpenMainWindow()
		}
	}()

	return nil
}

func (s *NoteService) SaveImage(noteId string, data []byte, filename string) (string, error) {
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	ext := filepath.Ext(filename)
	newFilename := id + ext

	_, err := s.db.Exec(`INSERT INTO images (id, note_id, filename, data) VALUES (?, ?, ?, ?)`,
		id, noteId, newFilename, data,
	)
	if err != nil {
		return "", err
	}
	return "/images/" + newFilename, nil
}

func (s *NoteService) GetStartupEnabled() bool {
	return IsStartupRegistered()
}

func (s *NoteService) ImageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/images/") {
			filename := strings.TrimPrefix(r.URL.Path, "/images/")

			var data []byte
			err := s.db.QueryRow(`SELECT data FROM images WHERE filename=?`, filename).Scan(&data)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			ext := filepath.Ext(filename)
			switch ext {
			case ".png":
				w.Header().Set("Content-Type", "image/png")
			case ".jpg", ".jpeg":
				w.Header().Set("Content-Type", "image/jpeg")
			case ".gif":
				w.Header().Set("Content-Type", "image/gif")
			case ".webp":
				w.Header().Set("Content-Type", "image/webp")
			}

			w.Write(data)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *NoteService) SetStartupEnabled(enabled bool) error {
	if enabled {
		return RegisterStartup()
	}
	return UnregisterStartup()
}

func (s *NoteService) loadSettings() {
	path := filepath.Join(os.Getenv("APPDATA"), "gamzamemo", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		s.settings = defaultSettings()
		return
	}
	json.Unmarshal(data, &s.settings)
}

func (s *NoteService) saveSettings() error {
	path := filepath.Join(os.Getenv("APPDATA"), "gamzamemo", "config.json")
	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *NoteService) GetSettings() Settings {
	return s.settings
}

func (s *NoteService) UpdateSettings(settings Settings) error {
	s.settings = settings
	// 모든 창에 설정 변경 알림
	s.app.Event.Emit("settings:updated", settings)
	return s.saveSettings()
}

func (s *NoteService) emitNotesUpdated() {
	notes, err := s.GetNotes()
	if err != nil {
		return
	}
	s.app.Event.Emit("notes:updated", notes)
}

func (s *NoteService) GetSystemFonts() ([]FontInfo, error) {
	return GetSystemFonts()
}

func (s *NoteService) OpenImageViewer(noteId string, src string) {
	x, y, width, height := 100, 100, 800, 600

	// 노트 창 위치 기반으로 중앙 계산
	if win, exists := s.windows[noteId]; exists {
		bounds := win.Bounds()
		x = bounds.X + (bounds.Width-width)/2
		y = bounds.Y + (bounds.Height-height)/2
	}

	win := s.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "이미지 보기",
		Width:     width,
		Height:    height,
		MinWidth:  300,
		MinHeight: 200,
		X:         x,
		Y:         y,
		URL:       "/?viewer&src=" + src,
		Frameless: true,
	})
	win.Show()
}
