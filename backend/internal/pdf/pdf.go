// Package pdf genera el certificado de beca.
package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/juan-camilo/becas-app/backend/assets"
)

type Datos struct {
	Codigo, Nombre, Club, Torneo string
	FechaNacimiento              time.Time
	Edad, Porcentaje             int
	ValorInscripcion, Descuento  string // ya formateados "$ 350.000"
	ValorAPagar                  string
	FechaLimite                  time.Time
	Emision                      time.Time
}

func fecha(t time.Time) string { return t.Format("02/01/2006") }

// Generar devuelve el PDF (carta, vertical) como bytes.
func Generar(d Datos) ([]byte, error) {
	p := fpdf.New("P", "mm", "Letter", "")
	tr := p.UnicodeTranslatorFromDescriptor("") // tildes y ñ en fuentes core
	p.SetMargins(20, 20, 20)
	p.SetAutoPageBreak(false, 0) // una sola hoja; el pie se ubica con SetY
	p.AddPage()

	// 1. Encabezado
	p.RegisterImageOptionsReader("logo", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(assets.Logo))
	p.ImageOptions("logo", 20, 18, 22, 22, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	p.SetXY(46, 24)
	p.SetFont("Helvetica", "B", 18)
	p.CellFormat(0, 10, tr("Academia Deportiva Horizonte"), "", 1, "L", false, 0, "")

	// 2. Título
	p.SetY(55)
	p.SetFont("Helvetica", "B", 22)
	p.CellFormat(0, 12, tr("Certificado de Beca Deportiva"), "", 1, "C", false, 0, "")
	p.Ln(8)

	seccion := func(titulo string) {
		p.SetFont("Helvetica", "B", 13)
		p.SetFillColor(238, 242, 255)
		p.CellFormat(0, 8, tr(titulo), "", 1, "L", true, 0, "")
		p.Ln(1)
	}
	fila := func(k, v string) {
		p.SetFont("Helvetica", "", 11)
		p.CellFormat(60, 7, tr(k), "", 0, "L", false, 0, "")
		p.SetFont("Helvetica", "B", 11)
		p.CellFormat(0, 7, tr(v), "", 1, "L", false, 0, "")
	}

	// 3. Deportista
	seccion("Datos del deportista")
	fila("Nombre", d.Nombre)
	fila("Fecha de nacimiento", fecha(d.FechaNacimiento))
	fila("Edad", fmt.Sprintf("%d años", d.Edad))
	fila("Club de origen", d.Club)
	p.Ln(4)

	// 4. Beca
	seccion("Datos de la beca")
	fila("Torneo", d.Torneo)
	fila("Porcentaje de beca", fmt.Sprintf("%d%%", d.Porcentaje))
	p.Ln(4)

	// 5. Valores
	seccion("Valores")
	fila("Valor de la inscripción", d.ValorInscripcion)
	fila("Descuento", d.Descuento)
	p.SetFillColor(220, 252, 231)
	p.SetFont("Helvetica", "B", 13)
	p.CellFormat(60, 10, tr("VALOR A PAGAR"), "", 0, "L", true, 0, "")
	p.CellFormat(0, 10, tr(d.ValorAPagar), "", 1, "L", true, 0, "")
	fila("Fecha límite de pago", fecha(d.FechaLimite))
	p.Ln(8)

	// 6. Código
	p.SetFont("Helvetica", "B", 16)
	p.CellFormat(0, 10, tr("Código de verificación: "+d.Codigo), "1", 1, "C", false, 0, "")
	p.SetFont("Helvetica", "", 10)
	p.CellFormat(0, 7, tr("Verifique este código en /api/verificar/"+d.Codigo), "", 1, "C", false, 0, "")

	// 7. Pie
	p.SetY(-25)
	p.SetFont("Helvetica", "I", 9)
	p.CellFormat(0, 5, tr("Fecha de emisión: "+fecha(d.Emision)), "", 1, "C", false, 0, "")
	p.CellFormat(0, 5, tr("Documento generado automáticamente — prototipo académico"), "", 1, "C", false, 0, "")

	var b bytes.Buffer
	if err := p.Output(&b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
