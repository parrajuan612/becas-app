package pdf

import (
	"bytes"
	"testing"
	"time"
)

func TestGenerarUnaPagina(t *testing.T) {
	b, err := Generar(Datos{Codigo: "BEC-2026-ABC234", Nombre: "Santiago Gómez Ñandú", Club: "Club Los Andes", Torneo: "Copa Horizonte 2027",
		FechaNacimiento: time.Date(2012, 3, 14, 0, 0, 0, 0, time.UTC), Edad: 14, Porcentaje: 50,
		ValorInscripcion: "$ 350.000", Descuento: "$ 175.000", ValorAPagar: "$ 175.000",
		FechaLimite: time.Date(2027, 2, 15, 0, 0, 0, 0, time.UTC), Emision: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(b, []byte("%PDF")) {
		t.Fatal("no es un PDF")
	}
	if n := bytes.Count(b, []byte("/Type /Page\n")); n != 1 {
		t.Errorf("el certificado debe tener 1 página, tiene %d", n)
	}
}
