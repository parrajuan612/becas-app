// Genera ejemplos/becas_200.xlsx (reproducible). Uso: go run ./cmd/generar-ejemplo [ruta]
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/juan-camilo/becas-app/backend/internal/ejemplo"
)

func main() {
	ruta := "../ejemplos/becas_200.xlsx"
	if len(os.Args) > 1 {
		ruta = os.Args[1]
	}
	if err := ejemplo.Generar(ruta); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generado:", ruta)
}
