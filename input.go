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
		}

	case tcell.KeyLeft:
		e.triggerBack()

	case tcell.KeyTab:
		if len(e.tabs) > 1 {
			e.activeTab = (e.activeTab + 1) % len(e.tabs)
		}

	case tcell.KeyCtrlW:
		if len(e.tabs) > 0 {
			e.tabs = append(e.tabs[:e.activeTab], e.tabs[e.activeTab+1:]...)
			if e.activeTab >= len(e.tabs) {
				e.activeTab = len(e.tabs) - 1
			}
		}

	case tcell.KeyF2:
		e.commandMode = true
		e.commandBuf = ""
	}
	return false
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
		content = []string{"Error al abrir el archivo."}
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
