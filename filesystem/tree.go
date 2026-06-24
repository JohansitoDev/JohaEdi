package filesystem

import (
	"os"
	"path/filepath"
)

type ArchivoNodo struct {
	Nombre       string
	EsDirectorio bool
	Ruta         string
	Hijos        []*ArchivoNodo
}

func LeerDirectorio(rutaBase string) (*ArchivoNodo, error) {
	info, err := os.Stat(rutaBase)
	if err != nil {
		return nil, err
	}

	raiz := &ArchivoNodo{
		Nombre:       filepath.Base(rutaBase),
		EsDirectorio: info.IsDir(),
		Ruta:         rutaBase,
	}

	if raiz.EsDirectorio {
		elementos, err := os.ReadDir(rutaBase)
		if err != nil {
			return raiz, nil
		}

		for _, elem := range elementos {
			if elem.Name()[0] == '.' && elem.Name() != "." {
				continue
			}
			rutaHijo := filepath.Join(rutaBase, elem.Name())
			nodoHijo, err := LeerDirectorio(rutaHijo)
			if err == nil {
				raiz.Hijos = append(raiz.Hijos, nodoHijo)
			}
		}
	}
	return raiz, nil
}

func FormatearArbol(nodo *ArchivoNodo, prefijo string) []string {
	var lineas []string
	if nodo == nil {
		return lineas
	}

	if nodo.EsDirectorio {
		lineas = append(lineas, prefijo+nodo.Nombre+"/")
		for i, hijo := range nodo.Hijos {
			ultimo := i == len(nodo.Hijos)-1
			nuevoPrefijo := prefijo + "  "
			if ultimo {
				nuevoPrefijo = prefijo + "  "
			}
			lineas = append(lineas, FormatearArbol(hijo, nuevoPrefijo)...)
		}
	} else {
		lineas = append(lineas, prefijo+nodo.Nombre)
	}
	return lineas
}

func ObtenerNodosPlanos(nodo *ArchivoNodo) []*ArchivoNodo {
	var nodos []*ArchivoNodo
	if nodo == nil {
		return nodos
	}
	nodos = append(nodos, nodo)
	for _, hijo := range nodo.Hijos {
		nodos = append(nodos, ObtenerNodosPlanos(hijo)...)
	}
	return nodos
}
