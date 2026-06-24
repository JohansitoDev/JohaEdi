package editor

import (
	"os"
	"strings"
)

type BufferEditor struct {
	Lineas       []string
	RutaArchivo  string
	CursorX      int
	CursorY      int
	Portapapeles string
}

func NuevoBuffer(ruta string) *BufferEditor {
	b := &BufferEditor{
		Lineas:      []string{""},
		RutaArchivo: ruta,
	}
	if ruta != "" {
		contenido, err := os.ReadFile(ruta)
		if err == nil {
			// Normalizar saltos de línea de Windows (\r\n) a (\n)
			texto := strings.ReplaceAll(string(contenido), "\r\n", "\n")
			b.Lineas = strings.Split(texto, "\n")
		}
	}
	return b
}

func (b *BufferEditor) GuardarCambios() error {
	if b.RutaArchivo == "" {
		b.RutaArchivo = "sin_titulo.txt"
	}
	contenido := strings.Join(b.Lineas, "\n")
	return os.WriteFile(b.RutaArchivo, []byte(contenido), 0644)
}

func (b *BufferEditor) CopiarLinea() {
	if b.CursorY >= 0 && b.CursorY < len(b.Lineas) {
		b.Portapapeles = b.Lineas[b.CursorY]
	}
}

func (b *BufferEditor) InsertarCaracter(r rune) {
	linea := b.Lineas[b.CursorY]
	nuevaLinea := linea[:b.CursorX] + string(r) + linea[b.CursorX:]
	b.Lineas[b.CursorY] = nuevaLinea
	b.CursorX++
}

// Nueva línea al presionar Enter
func (b *BufferEditor) InsertarSaltoLinea() {
	lineaActual := b.Lineas[b.CursorY]
	textoAntesCursor := lineaActual[:b.CursorX]
	textoDespuesCursor := lineaActual[b.CursorX:]

	b.Lineas[b.CursorY] = textoAntesCursor

	// Insertar la nueva línea justo abajo
	b.Lineas = append(b.Lineas[:b.CursorY+1], append([]string{textoDespuesCursor}, b.Lineas[b.CursorY+1:]...)...)
	b.CursorY++
	b.CursorX = 0
}

// Borrar caracteres con Backspace
func (b *BufferEditor) BorrarCaracter() {
	if b.CursorX > 0 {
		linea := b.Lineas[b.CursorY]
		b.Lineas[b.CursorY] = linea[:b.CursorX-1] + linea[b.CursorX:]
		b.CursorX--
	} else if b.CursorY > 0 {
		// Si está al inicio de la línea, sube el texto y lo concatena a la línea de arriba
		lineaActual := b.Lineas[b.CursorY]
		b.CursorY--
		b.CursorX = len(b.Lineas[b.CursorY])
		b.Lineas[b.CursorY] += lineaActual
		b.Lineas = append(b.Lineas[:b.CursorY+1], b.Lineas[b.CursorY+2:]...)
	}
}

func (b *BufferEditor) MoverArriba() {
	if b.CursorY > 0 {
		b.CursorY--
		if b.CursorX > len(b.Lineas[b.CursorY]) {
			b.CursorX = len(b.Lineas[b.CursorY])
		}
	}
}

func (b *BufferEditor) MoverAbajo() {
	if b.CursorY < len(b.Lineas)-1 {
		b.CursorY++
		if b.CursorX > len(b.Lineas[b.CursorY]) {
			b.CursorX = len(b.Lineas[b.CursorY])
		}
	}
}

func (b *BufferEditor) MoverIzquierda() {
	if b.CursorX > 0 {
		b.CursorX--
	} else if b.CursorY > 0 {
		b.CursorY--
		b.CursorX = len(b.Lineas[b.CursorY])
	}
}

func (b *BufferEditor) MoverDerecha() {
	if b.CursorX < len(b.Lineas[b.CursorY]) {
		b.CursorX++
	} else if b.CursorY < len(b.Lineas)-1 {
		b.CursorY++
		b.CursorX = 0
	}
}

func (b *BufferEditor) LimpiarBuffer(nuevaRuta string) {
	b.Lineas = []string{""}
	b.RutaArchivo = nuevaRuta
	b.CursorX = 0
	b.CursorY = 0
}
