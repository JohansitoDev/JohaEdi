package main

import (
	"fmt"
	"path/filepath"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

func (e *Editor) draw() {
	e.screen.Clear()
	w, h := e.screen.Size()
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorGreen)
	invStyle := tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack)

	mainH := h - e.terminalH - 2
	sidebarW := 28

	for y := 0; y < h-2; y++ {
		e.screen.SetContent(sidebarW, y, '│', nil, style)
	}

	dirTitle := fmt.Sprintf(" DIR: %s", filepath.Base(e.currentDir))
	e.drawString(0, 0, dirTitle, invStyle)

	for i, f := range e.files {
		if i+1 >= mainH {
			break
		}
		itemStyle := style
		if i == e.fileIndex && !e.editMode {
			itemStyle = invStyle
		}
		e.drawString(1, i+1, f, itemStyle)
	}

	mainX := sidebarW + 2
	maxLinesToDisplay := mainH - 2

	if len(e.tabs) > 0 {
		tabX := mainX
		for idx, t := range e.tabs {
			tabName := " " + t.Name + " "
			tabStyle := style
			if idx == e.activeTab {
				tabStyle = invStyle
			}
			e.drawString(tabX, 0, tabName, tabStyle)
			tabX += len(tabName) + 1
			e.screen.SetContent(tabX-1, 0, '│', nil, style)
		}

		active := e.tabs[e.activeTab]

		for idx := 0; idx < maxLinesToDisplay; idx++ {
			fileLineIdx := idx + e.textOffsetY
			if fileLineIdx >= len(active.Content) {
				break
			}
			e.drawString(mainX, idx+2, active.Content[fileLineIdx], style)
		}

		if e.editMode {
			visualCursorY := e.cursorY - e.textOffsetY
			e.screen.ShowCursor(mainX+e.cursorX, visualCursorY+2)
		} else {
			e.screen.HideCursor()
		}
	} else {
		e.screen.HideCursor()
		logo := []string{
			"██████  ██████  ██   ██  ██████  ███████ ██████  ██ ",
			"    ██ ██    ██ ██   ██ ██    ██ ██      ██   ██ ██ ",
			"    ██ ██    ██ ███████ ████████ █████   ██   ██ ██ ",
			"██  ██ ██    ██ ██   ██ ██    ██ ██      ██   ██ ██ ",
			" ████   ██████  ██   ██ ██    ██ ███████ ██████  ██ ",
			"                                                    ",
			"         ═╩═ JOHAEDI TERMINAL EDITOR ═╩═            ",
		}
		centerY := mainH / 3
		for idx, line := range logo {
			visualWidth := utf8.RuneCountInString(line)
			centerX := mainX + ((w - mainX - visualWidth) / 2)
			e.drawString(centerX, centerY+idx, line, style)
		}
	}

	termY := mainH + 1
	for x := 0; x < w; x++ {
		e.screen.SetContent(x, termY, '─', nil, style)
	}
	e.drawString(2, termY, " [ PANEL DE EJECUCIÓN & LOGS ] ", invStyle)

	logRow := termY + 1
	startLog := 0
	if len(e.outputLogs) > e.terminalH-1 {
		startLog = len(e.outputLogs) - (e.terminalH - 1)
	}
	for i := startLog; i < len(e.outputLogs); i++ {
		if logRow >= h-2 {
			break
		}
		e.drawString(1, logRow, e.outputLogs[i], style)
		logRow++
	}

	statusText := " [←/→] Moverse | [Tab] Cambiar Pestaña | [F2] Comando | [Enter] Editar | [Esc] Salir"
	if e.editMode {
		statusText = " MODO EDICIÓN: [Ctrl+S] Guardar Cambios | [F1] Volver al Explorador de Archivos"
	} else if e.commandMode {
		statusText = " MODO COMANDO: Ejecuta tareas en background (Ej: go run . , npm run dev, git status)"
	}
	e.drawString(0, h-2, fmt.Sprintf("%-*s", w, statusText), invStyle)

	if e.commandMode {
		e.drawString(0, h-1, "> "+e.commandBuf, style)
	} else {
		e.drawString(0, h-1, fmt.Sprintf("Ruta: %s", e.currentDir), style)
	}

	e.screen.Show()
}

func (e *Editor) drawString(x, y int, str string, style tcell.Style) {
	for _, r := range str {
		e.screen.SetContent(x, y, r, nil, style)
		x++
	}
}
