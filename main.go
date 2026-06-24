package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"unsafe"

	// <--- AGREGA ESTA LÍNEA
	"JohaEdi/editor"
	"JohaEdi/filesystem"

	"golang.org/x/term"
)

const (
	ColorVerde      = "\033[32m"
	ColorReset      = "\033[0m"
	LimpiarPantalla = "\033[H\033[2J"
	OcultarCursor   = "\033[?25l"
	MostrarCursor   = "\033[?25h"
)

var (
	focoEditor = true
	indexArbol = 0
)

// Forzar a Windows a interpretar secuencias ANSI nativamente de forma correcta
func activarAnsiWindows() {
	modkernel32 := syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleMode := modkernel32.NewProc("SetConsoleMode")
	procGetConsoleMode := modkernel32.NewProc("GetConsoleMode")

	handle, _ := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	var mode uint32

	// Usamos unsafe.Pointer de forma nativa como lo exige la API de Go
	procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	mode |= 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
	procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
}
func unsafePointer(v *uint32) uintptr {
	return uintptr(uintptr(unsafePointerRaw(v)))
}

func unsafePointerRaw(v *uint32) *uint32 {
	return v
}

func main() {
	// Activar soporte ANSI en Windows de forma nativa antes de renderizar nada
	activarAnsiWindows()

	rutaArchivoInicial := ""
	if len(os.Args) > 1 {
		rutaArchivoInicial = os.Args[1]
		focoEditor = true
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
	arbol, _ := filesystem.LeerDirectorio(".")

	ejecutarBucle(buf, arbol, oldState)
}

func ejecutarBucle(buf *editor.BufferEditor, arbol *filesystem.ArchivoNodo, oldState *term.State) {
	b := make([]byte, 3)
	for {
		ancho, alto, _ := term.GetSize(int(os.Stdout.Fd()))
		renderizarPantalla(buf, arbol, ancho, alto)

		n, err := os.Stdin.Read(b)
		if err != nil {
			break
		}

		if n == 1 {
			switch b[0] {
			case 3: // Ctrl + C
				buf.CopiarLinea()
			case 19: // Ctrl + S
				_ = buf.GuardarCambios()
			case 17: // Ctrl + Q
				return
			case 9: // TAB
				focoEditor = !focoEditor
			case 13: // ENTER
				if focoEditor {
					buf.InsertarSaltoLinea()
				} else {
					nodosPlanos := filesystem.ObtenerNodosPlanos(arbol)
					if indexArbol < len(nodosPlanos) && !nodosPlanos[indexArbol].EsDirectorio {
						buf = editor.NuevoBuffer(nodosPlanos[indexArbol].Ruta)
						focoEditor = true
						fmt.Print(LimpiarPantalla)
					}
				}
			case 127, 8: // BACKSPACE
				if focoEditor {
					buf.BorrarCaracter()
				}
			case 5: // Ctrl + E -> Modo Comando
				fmt.Print(MostrarCursor)
				term.Restore(int(os.Stdin.Fd()), oldState)

				ejecutarConsolaComandos(buf, alto)

				oldState, _ = term.MakeRaw(int(os.Stdin.Fd()))
				fmt.Print(OcultarCursor)
				fmt.Print(LimpiarPantalla)
				arbol, _ = filesystem.LeerDirectorio(".")
			default:
				if focoEditor && b[0] >= 32 && b[0] <= 126 {
					buf.InsertarCaracter(rune(b[0]))
				}
			}
		} else if n == 3 && b[0] == 27 && b[1] == 91 {
			nodosPlanos := filesystem.ObtenerNodosPlanos(arbol)
			switch b[2] {
			case 65: // Arriba
				if focoEditor {
					buf.MoverArriba()
				} else if indexArbol > 0 {
					indexArbol--
				}
			case 66: // Abajo
				if focoEditor {
					buf.MoverAbajo()
				} else if indexArbol < len(nodosPlanos)-1 {
					indexArbol++
				}
			case 68: // Izquierda
				if focoEditor {
					buf.MoverIzquierda()
				}
			case 67: // Derecha
				if focoEditor {
					buf.MoverDerecha()
				}
			}
		}
	}
}

func renderizarPantalla(buf *editor.BufferEditor, arbol *filesystem.ArchivoNodo, ancho, alto int) {
	fmt.Print("\033[H") // Redibujar estrictamente en la esquina superior izquierda (pantalla estática)
	fmt.Print(ColorVerde)

	lineasArbol := filesystem.FormatearArbol(arbol, "")

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

	for i := 0; i < alto-1; i++ {
		// --- PANEL IZQUIERDO (Lista limpia) ---
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
		fmt.Printf("%-25s   ", colIzquierda)

		// --- PANEL CENTRAL ---
		if len(buf.Lineas) == 1 && buf.Lineas[0] == "" && buf.RutaArchivo == "" {
			if i >= lineaInicioLogo && i < lineaInicioLogo+len(logo) {
				fmt.Print(logo[i-lineaInicioLogo])
			}
		} else {
			if i < len(buf.Lineas) {
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
		fmt.Print("\033[K\n") // Borrar residuos de la derecha de forma limpia sin saltos locos
	}

	// --- BARRA INFERIOR ---
	fmt.Printf("\033[%d;1H\033[K", alto)
	seccionActual := "EDITOR"
	if !focoEditor {
		seccionActual = "EXPLORADOR"
	}
	archivoActivo := buf.RutaArchivo
	if archivoActivo == "" {
		archivoActivo = "Ninguno"
	}
	fmt.Printf("[TAB] %s | Archivo: %s | [Ctrl+S] Guardar | [Ctrl+E] Terminal | [Ctrl+Q] Salir", seccionActual, archivoActivo)
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

	partes := strings.Fields(comandoInput)
	comandoPrincipal := partes[0]

	fmt.Print(LimpiarPantalla + "\033[H")

	var cmd *exec.Cmd
	if len(partes) > 1 {
		cmd = exec.Command("cmd", "/c", comandoPrincipal, strings.Join(partes[1:], " "))
	} else {
		cmd = exec.Command("cmd", "/c", comandoPrincipal)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	_ = cmd.Run()

	fmt.Println("\n\n[Presiona Enter para regresar a JohaEdi]")
	_, _ = reader.ReadString('\n')
}
