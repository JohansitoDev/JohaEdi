package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (e *Editor) handleInput(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlQ:
		return true

	case tcell.KeyUp:
		if e.fileIndex > 0 {
			e.fileIndex--
		}

	case tcell.KeyDown:
		if e.fileIndex < len(e.files)-1 {
			e.fileIndex++
		}

	case tcell.KeyEnter:

		if len(e.files) == 0 {
			return false
		}
		selected := e.files[e.fileIndex]

		if selected == ".. (Volver)" {
			e.currentDir = filepath.Dir(e.currentDir)
			e.fileIndex = 0
			e.updateFileList()
			e.currentFile = ""
		} else if strings.HasPrefix(selected, "📁 ") {
			dirName := strings.TrimPrefix(selected, "📁 ")
			e.currentDir = filepath.Join(e.currentDir, dirName)
			e.fileIndex = 0
			e.updateFileList()
			e.currentFile = ""
		} else if strings.HasPrefix(selected, "📄 ") {
			fileName := strings.TrimPrefix(selected, "📄 ")
			e.currentFile = filepath.Join(e.currentDir, fileName)

			data, err := os.ReadFile(e.currentFile)
			if err == nil {
				e.fileContent = strings.Split(string(data), "\n")
			} else {
				e.fileContent = []string{"Error al leer el archivo."}
			}
		}

	case tcell.KeyF1:

		e.commandMode = true
		e.commandBuf = "mkdir "

	case tcell.KeyF2:

		e.commandMode = true
		e.commandBuf = ""
	}
	return false
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

		if parts[0] == "mkdir" && len(parts) > 1 {
			targetDir := filepath.Join(e.currentDir, parts[1])
			os.MkdirAll(targetDir, 0755)
			e.outputLog = "Carpeta creada: " + parts[1]
			e.updateFileList()
			return
		}

		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = e.currentDir

		out, err := cmd.CombinedOutput()
		if err != nil {
			e.outputLog = fmt.Sprintf("Error: %s", string(out))
		} else {
			e.outputLog = fmt.Sprintf("Éxito: %s", string(out))
		}

		e.updateFileList()

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.commandBuf) > 0 {
			e.commandBuf = e.commandBuf[:len(e.commandBuf)-1]
		}

	case tcell.KeyRune:
		e.commandBuf += string(ev.Rune())
	}
}
