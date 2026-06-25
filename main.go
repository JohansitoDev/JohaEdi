package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
)

type Editor struct {
	screen      tcell.Screen
	currentDir  string
	files       []string
	fileIndex   int
	currentFile string
	fileContent []string
	commandMode bool
	commandBuf  string
	outputLog   string
}

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%v", err)
	}

	s.Clear()

	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorGreen)
	s.SetStyle(defStyle)

	home, _ := os.UserHomeDir()
	desktopPath := filepath.Join(home, "Desktop")
	if _, err := os.Stat(desktopPath); err != nil {
		desktopPath = home
	}

	e := &Editor{
		screen:     s,
		currentDir: desktopPath,
	}

	e.updateFileList()

	for {
		e.draw()

		ev := e.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			e.screen.Sync()
		case *tcell.EventKey:
			if e.commandMode {
				e.handleCommandInput(ev)
			} else {
				if e.handleInput(ev) {
					e.screen.Fini()
					return
				}
			}
		}
	}
}

func (e *Editor) updateFileList() {
	e.files = []string{".. (Volver)"}
	entries, err := os.ReadDir(e.currentDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		prefix := "📄 "
		if entry.IsDir() {
			prefix = "📁 "
		}
		e.files = append(e.files, prefix+entry.Name())
	}
	if e.fileIndex >= len(e.files) {
		e.fileIndex = 0
	}
}
