package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (e *Editor) handleInput(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		return true

	case tcell.KeyUp:
		if e.fileIndex > 0 {
			e.fileIndex--
		}

	case tcell.KeyDown:
		if e.fileIndex < len(e.files)-1 {
			e.fileIndex++
		}

	case tcell.KeyRight, tcell.KeyEnter:
		if len(e.files) == 0 {
			return false
		}
		selected := e.files[e.fileIndex]

		if selected == ".. (Volver)" {
			e.triggerBack()
		} else if strings.HasPrefix(selected, "📁 ") {
			dirName := strings.TrimPrefix(selected, "📁 ")
			e.currentDir = filepath.Join(e.currentDir, dirName)
			e.fileIndex = 0
			e.updateFileList()
		} else if strings.HasPrefix(selected, "📄 ") {
			fileName := strings.TrimPrefix(selected, "📄 ")
			fullPath := filepath.Join(e.currentDir, fileName)
			e.openFileInTab(fileName, fullPath)
			e.editMode = true
			e.cursorX = 0
			e.cursorY = 0
			e.textOffsetY = 0
		}

	case tcell.KeyLeft:
		e.triggerBack()

	case tcell.KeyTab:
		if len(e.tabs) > 1 {
			e.activeTab = (e.activeTab + 1) % len(e.tabs)
			e.cursorX = 0
			e.cursorY = 0
			e.textOffsetY = 0
		}

	case tcell.KeyCtrlW:
		if len(e.tabs) > 0 {
			e.tabs = append(e.tabs[:e.activeTab], e.tabs[e.activeTab+1:]...)
			if e.activeTab >= len(e.tabs) {
				e.activeTab = len(e.tabs) - 1
			}
			if len(e.tabs) == 0 {
				e.editMode = false
			}
			e.cursorX = 0
			e.cursorY = 0
			e.textOffsetY = 0
		}

	case tcell.KeyF2:
		e.commandMode = true
		e.commandBuf = ""
	}
	return false
}

func (e *Editor) handleEditInput(ev *tcell.EventKey) {
	if len(e.tabs) == 0 {
		e.editMode = false
		return
	}

	_, h := e.screen.Size()
	mainH := h - e.terminalH - 2
	maxLinesToDisplay := mainH - 2

	active := &e.tabs[e.activeTab]

	switch ev.Key() {
	case tcell.KeyF1:
		e.editMode = false

	case tcell.KeyCtrlS:
		fullContent := strings.Join(active.Content, "\n")
		err := os.WriteFile(active.Path, []byte(fullContent), 0644)
		if err == nil {
			e.outputLogs = append(e.outputLogs, "💾 Archivo guardado: "+active.Name)
		} else {
			e.outputLogs = append(e.outputLogs, "❌ Error al guardar: "+err.Error())
		}

	case tcell.KeyUp:
		if e.cursorY > 0 {
			e.cursorY--
			if e.cursorY < e.textOffsetY {
				e.textOffsetY = e.cursorY
			}
			if e.cursorX > len(active.Content[e.cursorY]) {
				e.cursorX = len(active.Content[e.cursorY])
			}
		}

	case tcell.KeyDown:
		if e.cursorY < len(active.Content)-1 {
			e.cursorY++
			if e.cursorY >= e.textOffsetY+maxLinesToDisplay {
				e.textOffsetY++
			}
			if e.cursorX > len(active.Content[e.cursorY]) {
				e.cursorX = len(active.Content[e.cursorY])
			}
		}

	case tcell.KeyLeft:
		if e.cursorX > 0 {
			e.cursorX--
		} else if e.cursorY > 0 {
			e.cursorY--
			if e.cursorY < e.textOffsetY {
				e.textOffsetY = e.cursorY
			}
			e.cursorX = len(active.Content[e.cursorY])
		}

	case tcell.KeyRight:
		if e.cursorX < len(active.Content[e.cursorY]) {
			e.cursorX++
		} else if e.cursorY < len(active.Content)-1 {
			e.cursorY++
			if e.cursorY >= e.textOffsetY+maxLinesToDisplay {
				e.textOffsetY++
			}
			e.cursorX = 0
		}

	case tcell.KeyEnter:
		currentLine := active.Content[e.cursorY]
		leftPart := currentLine[:e.cursorX]
		rightPart := currentLine[e.cursorX:]

		active.Content[e.cursorY] = leftPart
		active.Content = append(active.Content[:e.cursorY+1], append([]string{rightPart}, active.Content[e.cursorY+1:]...)...)
		e.cursorY++

		if e.cursorY >= e.textOffsetY+maxLinesToDisplay {
			e.textOffsetY++
		}
		e.cursorX = 0

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if e.cursorX > 0 {
			currentLine := active.Content[e.cursorY]
			active.Content[e.cursorY] = currentLine[:e.cursorX-1] + currentLine[e.cursorX:]
			e.cursorX--
		} else if e.cursorY > 0 {
			prevLineLen := len(active.Content[e.cursorY-1])
			active.Content[e.cursorY-1] += active.Content[e.cursorY]
			active.Content = append(active.Content[:e.cursorY], active.Content[e.cursorY+1:]...)
			e.cursorY--

			if e.cursorY < e.textOffsetY {
				e.textOffsetY = e.cursorY
			}
			e.cursorX = prevLineLen
		}

	case tcell.KeyRune:
		currentLine := active.Content[e.cursorY]
		active.Content[e.cursorY] = currentLine[:e.cursorX] + string(ev.Rune())
		e.cursorX++
	}
}

func (e *Editor) triggerBack() {
	e.currentDir = filepath.Dir(e.currentDir)
	e.fileIndex = 0
	e.updateFileList()
}

func (e *Editor) openFileInTab(name, path string) {
	for idx, t := range e.tabs {
		if t.Path == path {
			e.activeTab = idx
			return
		}
	}

	data, err := os.ReadFile(path)
	var content []string
	if err == nil {
		content = strings.Split(string(data), "\n")
	} else {
		content = []string{""}
	}

	newTab := Tab{Name: name, Path: path, Content: content}
	e.tabs = append(e.tabs, newTab)
	e.activeTab = len(e.tabs) - 1
}

func (e *Editor) handleCommandInput(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		e.commandMode = false
		e.commandBuf = ""

	case tcell.KeyEnter:
		e.commandMode = false
		cmdStr := strings.TrimSpace(e.commandBuf)
		if cmdStr == "" {
			return
		}

		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			return
		}

		e.outputLogs = append(e.outputLogs, "🚀 Exec: "+cmdStr)

		go func() {
			cmd := exec.Command(parts[0], parts[1:]...)
			cmd.Dir = e.currentDir

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return
			}
			cmd.Stderr = cmd.Stdout

			if err := cmd.Start(); err != nil {
				e.outputLogs = append(e.outputLogs, "❌ Error: "+err.Error())
				return
			}

			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				e.outputLogs = append(e.outputLogs, scanner.Text())
			}
			cmd.Wait()
			e.updateFileList()
		}()

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.commandBuf) > 0 {
			e.commandBuf = e.commandBuf[:len(e.commandBuf)-1]
		}

	case tcell.KeyRune:
		e.commandBuf += string(ev.Rune())
	}
}
