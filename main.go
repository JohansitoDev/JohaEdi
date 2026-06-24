package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"JohaEdi/editor"
	"JohaEdi/filesystem"

	"golang.org/x/term"
)

const (
	ColorVerde      = "\033[32m"
	ColorGris       = "\033[90m"
	ColorReset      = "\033[0m"
	LimpiarPantalla = "\033[H\033[2J"
	OcultarCursor   = "\033[?25l"
	MostrarCursor   = "\033[?25h"
)

var (
	focoEditor = false
	indexArbol = 0
)

func activarAnsiWindows() {
	modkernel32 := syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleMode := modkernel32.NewProc("SetConsoleMode")
	procGetConsoleMode := modkernel32.NewProc("GetConsoleMode")

	handle, _ := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	var mode uint32

	procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	mode |= 0x0004
	procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
}

func filtrarNodosInternos(nodo *filesystem.ArchivoNodo) *filesystem.ArchivoNodo {
	if nodo == nil {
		return nil
	}
	var nuevosHijos []*filesystem.ArchivoNodo
	for _, hijo := range nodo.Hijos {
		nombreLower := strings.ToLower(hijo.Nombre)
		if nombreLower == "editor" || nombreLower == "filesystem" || nombreLower == "ui" ||
			nombreLower == "go.mod" || nombreLower == "go.sum" || nombreLower == "main.go" ||
			strings.HasPrefix(nombreLower, "johaedi.exe") || nombreLower == "readme.md" ||
			(len(hijo.Nombre) > 0 && hijo.Nombre[0] == '.') {
			continue
		}
		nuevosHijos = append(nuevosHijos, filtrarNodosInternos(hijo))
	}
	nodo.Hijos = nuevosHijos
	return nodo
}

func main() {
	activarAnsiWindows()

	rutaArchivoInicial := ""
	if len(os.Args) > 1 {
		fi, err := os.Stat(os.Args[1])
		if err == nil {
			if fi.IsDir() {
				_ = os.Chdir(os.Args[1])
			} else {
				_ = os.Chdir(filepath.Dir(os.Args[1]))
				rutaArchivoInicial = filepath.Base(os.Args[1])
				focoEditor = true
			}
		}
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Error al iniciar modo terminal raw:", err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		term.Restore(int(os.Stdin.Fd()), oldState)
		fmt.Print(ColorReset + LimpiarPantalla + MostrarCursor)
		os.Exit(0)
	}()

	fmt.Print(OcultarCursor)
	defer fmt.Print(MostrarCursor)
	fmt.Print(LimpiarPantalla)

	buf := editor.NuevoBuffer(rutaArchivoInicial)
	dirActual, _ := os.Getwd()
	arbolRaw, _ := filesystem.LeerDirectorio(dirActual)
	arbol := filtrarNodosInternos(arbolRaw)

	b := make([]byte, 3)
	for {
		ancho, alto, _ := term.GetSize(int(os.Stdout.Fd()))
		renderizarPantalla(buf, arbol, ancho, alto)

		n, err := os.Stdin.Read(b)
		if err != nil {
			break
		}

		var nodosPlanos []*filesystem.ArchivoNodo
		if arbol != nil {
			nodosPlanos = filesystem.ObtenerNodosPlanos(arbol)
		}

		if n == 1 {
			switch b[0] {
			case 9: // TAB -> Cambiar entre Explorador y Editor
				focoEditor = !focoEditor
			case 19: // Ctrl + S -> Guardar Cambios del archivo abierto
				_ = buf.GuardarCambios()
			case 17: // Ctrl + Q -> Salir
				return
			case 13: // ENTER
				if focoEditor {
					buf.InsertarSaltoLinea()
				} else {
					if len(nodosPlanos) > 0 && indexArbol < len(nodosPlanos) {
						nodoSeleccionado := nodosPlanos[indexArbol]
						if nodoSeleccionado.EsDirectorio {
							_ = os.Chdir(nodoSeleccionado.Ruta)
							indexArbol = 0
							fmt.Print(LimpiarPantalla)
							dirActual, _ = os.Getwd()
							arbolRaw, _ = filesystem.LeerDirectorio(dirActual)
							arbol = filtrarNodosInternos(arbolRaw)
						} else {
							buf = editor.NuevoBuffer(nodoSeleccionado.Ruta)
							focoEditor = true
							fmt.Print(LimpiarPantalla)
						}
					}
				}
			case 127, 8: // BACKSPACE
				if focoEditor {
					buf.BorrarCaracter()
				}
			case 114, 82: // 'R' -> Crear Archivo
				if !focoEditor {
					term.Restore(int(os.Stdin.Fd()), oldState)
					fmt.Printf("\033[%d;1H\033[KNombre del nuevo archivo: ", alto)
					reader := bufio.NewReader(os.Stdin)
					nombre, _ := reader.ReadString('\n')
					nombre = strings.TrimSpace(nombre)

					if nombre != "" {
						_ = os.WriteFile(nombre, []byte(""), 0644)
					}
					oldState, _ = term.MakeRaw(int(os.Stdin.Fd()))
					fmt.Print(LimpiarPantalla)
					dirActual, _ = os.Getwd()
					arbolRaw, _ = filesystem.LeerDirectorio(dirActual)
					arbol = filtrarNodosInternos(arbolRaw)
				} else {
					buf.InsertarCaracter(rune(b[0]))
				}
			case 116, 84: // 'T' -> Crear Carpeta
				if !focoEditor {
					term.Restore(int(os.Stdin.Fd()), oldState)
					fmt.Printf("\033[%d;1H\033[KNombre de la nueva carpeta: ", alto)
					reader := bufio.NewReader(os.Stdin)
					nombre, _ := reader.ReadString('\n')
					nombre = strings.TrimSpace(nombre)

					if nombre != "" {
						_ = os.MkdirAll(nombre, 0755)
					}
					oldState, _ = term.MakeRaw(int(os.Stdin.Fd()))
					fmt.Print(LimpiarPantalla)
					dirActual, _ = os.Getwd()
					arbolRaw, _ = filesystem.LeerDirectorio(dirActual)
					arbol = filtrarNodosInternos(arbolRaw)
				} else {
					buf.InsertarCaracter(rune(b[0]))
				}
			case 5: // Ctrl + E -> Terminal
				fmt.Print(MostrarCursor)
				term.Restore(int(os.Stdin.Fd()), oldState)

				ejecutarConsolaComandos(buf, alto)

				oldState, _ = term.MakeRaw(int(os.Stdin.Fd()))
				fmt.Print(OcultarCursor)
				fmt.Print(LimpiarPantalla)
				dirActual, _ = os.Getwd()
				arbolRaw, _ = filesystem.LeerDirectorio(dirActual)
				arbol = filtrarNodosInternos(arbolRaw)
			default:
				if focoEditor && b[0] >= 32 && b[0] <= 126 {
					buf.InsertarCaracter(rune(b[0]))
				}
			}
		} else if n == 3 && b[0] == 27 && b[1] == 91 { // MOVIMIENTO FÍSICO CORREGIDO
			switch b[2] {
			case 65: // Flecha Arriba
				if focoEditor {
					buf.MoverArriba() // Ahora sube correctamente de línea de código
				} else if indexArbol > 0 {
					indexArbol--
				}
			case 66: // Flecha Abajo
				if focoEditor {
					buf.MoverAbajo() // Ahora baja correctamente a la siguiente línea
				} else if len(nodosPlanos) > 0 && indexArbol < len(nodosPlanos)-1 {
					indexArbol++
				}
			case 68: // Flecha Izquierda
				if focoEditor {
					buf.MoverIzquierda()
				}
			case 67: // Flecha Derecha
				if focoEditor {
					buf.MoverDerecha()
				}
			}
		}
	}
}

func renderizarPantalla(buf *editor.BufferEditor, arbol *filesystem.ArchivoNodo, ancho, alto int) {
	fmt.Print("\033[H")
	fmt.Print(ColorVerde)

	var lineasArbol []string
	if arbol != nil && len(arbol.Hijos) > 0 {
		lineasArbol = filesystem.FormatearArbol(arbol, "")
	}

	logo := []string{
		"██╗ ██████╗ ██╗  ██╗ █████╗ ███████╗██████╗ ██╗",
		"██║██╔═══██╗██║  ██║██╔══██╗██╔════╝██╔══██╗██║",
		"██║██║   ██║███████║███████║█████╗  ██║  ██║██║",
		"██╗██║██║   ██║██╔══██║██╔══██║██╔══╝  ██║  ██║██║",
		"╚█████╔╝╚██████╔╝██║  ██║██║  ██║███████╗██████╔╝ ██║",
		" ╚════╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═════╝  ╚═╝",
		"                                                     ",
		"               -- MINIMALIST CODE EDITOR --          ",
	}

	lineaInicioLogo := (alto - len(logo)) / 2
	if lineaInicioLogo < 0 {
		lineaInicioLogo = 0
	}

	tieneArchivosReales := len(lineasArbol) > 0
	tieneTextoAbierto := !(len(buf.Lineas) == 1 && buf.Lineas[0] == "" && buf.RutaArchivo == "")

	for i := 0; i < alto-1; i++ {
		// --- PANEL IZQUIERDO ---
		colIzquierda := ""
		if i < len(lineasArbol) {
			colIzquierda = lineasArbol[i]
			if !focoEditor && i == indexArbol {
				colIzquierda = "> " + colIzquierda
			} else {
				colIzquierda = "  " + colIzquierda
			}
		}

		if len(colIzquierda) > 25 {
			colIzquierda = colIzquierda[:25]
		}

		if tieneArchivosReales {
			fmt.Printf("%-25s │ ", colIzquierda)
		} else {
			fmt.Printf("%-25s   ", colIzquierda)
		}

		// --- PANEL DERECHO (SCROLL ESTÁTICO) ---
		if !tieneArchivosReales && !tieneTextoAbierto {
			if i >= lineaInicioLogo && i < lineaInicioLogo+len(logo) {
				fmt.Print(logo[i-lineaInicioLogo])
			}
		} else {
			if i < len(buf.Lineas) && i < (alto-2) {
				// Imprimir número de línea fijo a la izquierda
				fmt.Printf("%s%3d  %s", ColorGris, i+1, ColorVerde)
				lineaGrafica := buf.Lineas[i]

				if focoEditor && i == buf.CursorY {
					if buf.CursorX < len(lineaGrafica) {
						fmt.Print(lineaGrafica[:buf.CursorX] + "█" + lineaGrafica[buf.CursorX+1:])
					} else {
						fmt.Print(lineaGrafica + "█")
					}
				} else {
					fmt.Print(lineaGrafica)
				}
			}
		}
		fmt.Print("\033[K\n")
	}

	// --- BARRA INFERIOR ---
	fmt.Printf("\033[%d;1H\033[K", alto)
	seccionActual := "EXPLORADOR"
	if focoEditor {
		seccionActual = "EDITOR"
	}
	archivoActivo := filepath.Base(buf.RutaArchivo)
	if buf.RutaArchivo == "" {
		archivoActivo = "Ninguno"
	}

	if !focoEditor {
		fmt.Printf("[TAB] Modo: %s | [R] Nuevo Archivo | [T] Nueva Carpeta | [Ctrl+E] Terminal | [Ctrl+Q] Salir", seccionActual)
	} else {
		fmt.Printf("[TAB] Modo: %s | Archivo: %s | [Ctrl+S] Guardar | [Ctrl+E] Terminal | [Ctrl+Q] Salir", seccionActual, archivoActivo)
	}
}

func ejecutarConsolaComandos(buf *editor.BufferEditor, alto int) {
	fmt.Printf("\033[%d;1H\033[K", alto)
	fmt.Print("JohaEdi-Terminal> ")

	reader := bufio.NewReader(os.Stdin)
	comandoInput, _ := reader.ReadString('\n')
	comandoInput = strings.TrimSpace(comandoInput)

	if comandoInput == "" {
		return
	}

	fmt.Print(LimpiarPantalla + "\033[H")

	cmd := exec.Command("cmd", "/c", comandoInput)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	_ = cmd.Run()

	fmt.Println("\n\n[Presiona Enter para regresar a JohaEdi]")
	_, _ = reader.ReadString('\n')
}
