// Package excel genera la plantilla y lee los archivos de carga.
package excel

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"

	"github.com/juan-camilo/becas-app/backend/internal/service"
)

const Hoja = "becas"

var Columnas = []string{"nombre", "numero", "correo", "club", "fecha_nacimiento", "torneo", "porcentaje_beca", "autoriza"}

// Plantilla devuelve el .xlsx oficial con encabezados y una fila de ejemplo.
func Plantilla() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	f.SetSheetName("Sheet1", Hoja)
	ejemplo := []any{"Santiago Gómez Ruiz", "3001234567", "acudiente@gmail.com", "Club Los Andes",
		"14/03/2012", "Copa Horizonte 2027", 50, "SI"}
	for i, c := range Columnas {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(Hoja, cell, c)
		cell, _ = excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(Hoja, cell, ejemplo[i])
	}
	f.SetColWidth(Hoja, "A", "H", 22)
	var b bytes.Buffer
	if err := f.Write(&b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Leer lee la hoja `becas` (o la primera). Rechaza el archivo si los encabezados no son exactamente los esperados.
// Las filas totalmente vacías se ignoran; el número de fila conserva el del Excel.
func Leer(r io.Reader) ([]service.FilaCruda, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("el archivo no es un .xlsx válido")
	}
	defer f.Close()
	hoja := Hoja
	if idx, _ := f.GetSheetIndex(Hoja); idx < 0 {
		hoja = f.GetSheetName(0)
	}
	// RawCellValue evita que las fechas se conviertan a texto con formato regional.
	filas, err := f.GetRows(hoja, excelize.Options{RawCellValue: true})
	if err != nil || len(filas) == 0 {
		return nil, fmt.Errorf("el archivo está vacío")
	}

	pos := map[string]int{}
	var desconocidos []string
	for i, h := range filas[0] {
		h = strings.ToLower(strings.TrimSpace(h))
		if h == "" {
			continue
		}
		known := false
		for _, c := range Columnas {
			if c == h {
				known = true
			}
		}
		if !known {
			desconocidos = append(desconocidos, h)
		} else {
			pos[h] = i
		}
	}
	var faltan []string
	for _, c := range Columnas {
		if _, ok := pos[c]; !ok {
			faltan = append(faltan, c)
		}
	}
	if len(faltan) > 0 || len(desconocidos) > 0 {
		msg := "Encabezados incorrectos."
		if len(faltan) > 0 {
			msg += " Faltan: " + strings.Join(faltan, ", ") + "."
		}
		if len(desconocidos) > 0 {
			msg += " Desconocidos: " + strings.Join(desconocidos, ", ") + "."
		}
		return nil, fmt.Errorf("%s Columnas esperadas: %s", msg, strings.Join(Columnas, ", "))
	}

	var out []service.FilaCruda
	for i, row := range filas[1:] {
		get := func(c string) string {
			if p := pos[c]; p < len(row) {
				return strings.TrimSpace(row[p])
			}
			return ""
		}
		fc := service.FilaCruda{
			Fila: i + 2, Nombre: get("nombre"), Numero: get("numero"), Correo: get("correo"), Club: get("club"),
			FechaNac: get("fecha_nacimiento"), Torneo: get("torneo"), Porcentaje: get("porcentaje_beca"), Autoriza: get("autoriza"),
		}
		if fc == (service.FilaCruda{Fila: fc.Fila}) {
			continue
		}
		out = append(out, fc)
	}
	return out, nil
}
