package excel

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestPlantillaSeLee(t *testing.T) {
	b, err := Plantilla()
	if err != nil {
		t.Fatal(err)
	}
	filas, err := Leer(bytes.NewReader(b))
	if err != nil || len(filas) != 1 {
		t.Fatalf("filas=%d err=%v", len(filas), err)
	}
	if filas[0].Fila != 2 || filas[0].Nombre != "Santiago Gómez Ruiz" || filas[0].Porcentaje != "50" {
		t.Errorf("fila de ejemplo mal leída: %+v", filas[0])
	}
}

func xlsx(t *testing.T, filas [][]any) *bytes.Reader {
	t.Helper()
	f := excelize.NewFile()
	for i, fila := range filas {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		f.SetSheetRow("Sheet1", cell, &fila)
	}
	var b bytes.Buffer
	f.Write(&b)
	return bytes.NewReader(b.Bytes())
}

func TestEncabezadosIncorrectos(t *testing.T) {
	_, err := Leer(xlsx(t, [][]any{{"nombre", "numero", "extra"}}))
	if err == nil || !strings.Contains(err.Error(), "Faltan") || !strings.Contains(err.Error(), "Desconocidos: extra") {
		t.Fatalf("error inesperado: %v", err)
	}
}

func TestFilasVaciasSeIgnoran(t *testing.T) {
	enc := []any{"NOMBRE", "Numero", "correo", "club", "fecha_nacimiento", "torneo", "porcentaje_beca", "autoriza"}
	fila := []any{"Ana Ruiz", "3001112233", "a@b.co", "Club", "01/01/2012", "T", 10, "SI"}
	filas, err := Leer(xlsx(t, [][]any{enc, {}, fila}))
	if err != nil || len(filas) != 1 || filas[0].Fila != 3 {
		t.Fatalf("filas=%+v err=%v", filas, err)
	}
}
