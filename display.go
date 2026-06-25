package main

import (
	"fmt"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
)

func (e *Editor) draw() {
	e.screen.Clear()
	w, h := e.screen.Size()
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorGreen)
	invStyle := tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack)

	// 1. DIBUJAR EXPLORADOR DE ARCHIVOS (Lateral izquierdo - 25 columnas)
	sidebarW := 28
	for y := 0; y < h-2; y++ {
		e.screen.SetContent(sidebarW, y, '│', nil, style)
	}

	// Mostrar ruta actual recortada si es muy larga
	dirTitle := fmt.Sprintf(" DIR: %s", filepath.Base(e.currentDir))
	e.drawString(0, 0, dirTitle, invStyle)

	for i, f := range e.files {
		if i+1 >= h-2 {
			break
		}
		itemStyle := style
		if i == e.fileIndex {
			itemStyle = invStyle // Resaltar seleccionado
		}
		e.drawString(1, i+1, f, itemStyle)
	}

	// 2. DIBUJAR ÁREA PRINCIPAL
	mainX := sidebarW + 2
	if e.currentFile == "" {
		// PANTALLA DE BIENVENIDA: Logo centralizado si no hay archivo abierto
		logo := []string{
			"   ║║║║║║║║║║║║║║║║║║║║║║║║║   ",
			"   ║  ║║ ╔═══╗ ║  ║  ║ ╔═══╗ ║   ",
			"   ║  ║║ ║   ║ ║══╣  ║ ║   ║ ║   ",
			"   ╚══╝║ ╚═══╝ ╩  ╩══╝ ╚═══╝ ╩   ",
			"      ═╩═ JOHAEDI EDITOR ═╩═     ",
		}
		centerY := h / 3
		for idx, line := range logo {
			centerX := mainX + ((w - mainX - len(line)) / 2)
			e.drawString(centerX, centerY+idx, line, style)
		}

		// Mostrar logs de comandos abajo si existen
		if e.outputLog != "" {
			e.drawString(mainX, h-4, "═══ Última Salida ═══", style)
			e.drawString(mainX, h-3, e.outputLog, style)
		}
	} else {
		// Mostrar contenido del archivo abierto
		e.drawString(mainX, 0, fmt.Sprintf("📄 Leyendo: %s", filepath.Base(e.currentFile)), invStyle)
		for idx, line := range e.fileContent {
			if idx+2 >= h-2 {
				break
			}
			e.drawString(mainX, idx+2, line, style)
		}
	}

	// 3. BARRA DE ESTADO INFERIOR
	statusText := " [F1] Crear Carpeta | [F2] Comando Terminal | [Enter] Abrir/Entrar | [Esc/Ctrl+Q] Salir"
	if e.commandMode {
		statusText = " MODO COMANDO: Escribe tu instrucción (Ej: git status / npm init) y presiona Enter"
	}
	e.drawString(0, h-2, fmt.Sprintf("%-*s", w, statusText), invStyle)

	// Línea de Comandos activa
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
