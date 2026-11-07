// Package ejemplo genera el Excel de ejemplo (200 filas) de forma reproducible.
package ejemplo

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/xuri/excelize/v2"
)

var nombres = []string{"Santiago", "Valentina", "Mateo", "Sofía", "Samuel", "Isabella", "Emmanuel", "Mariana", "Juan José", "Salomé",
	"Thiago", "Luciana", "Nicolás", "Gabriela", "Andrés", "Daniela", "Felipe", "Camila", "Sebastián", "María Paula",
	"Tomás", "Antonella", "Alejandro", "Sara", "David", "Manuela", "Esteban", "Juliana", "Martín", "Ángela"}

var apellidos = []string{"Gómez", "Rodríguez", "Martínez", "López", "García", "Hernández", "Ramírez", "Torres", "Díaz", "Vargas",
	"Castro", "Ortiz", "Rojas", "Moreno", "Muñoz", "Parra", "Quintero", "Agudelo", "Cárdenas", "Peña"}

var clubes = []string{"Club Los Andes", "Club Bahía Sur", "Club Estrella del Norte", "Club Deportivo Cafetero", "Club Nueva Generación"}
var torneos = []string{"Copa Horizonte 2027", "Torneo Interclubes Sub-15", "Festival Infantil de Verano"}

type fila struct {
	nombre, numero, correo, club, fecha, torneo string
	pct                                         int
	autoriza                                    string
	fechaExcel                                  *time.Time // si no es nil, se escribe como fecha nativa de Excel
}

// Composición exacta (SDD §10.2).
const (
	Validas, NumeroInvalido, CorreoInvalido, PctFuera, TorneoInexistente, SinAutorizacion, Duplicadas = 170, 8, 5, 3, 4, 5, 5
)

// Generar escribe el Excel en ruta. Misma semilla => mismo archivo.
func Generar(ruta string) error {
	r := rand.New(rand.NewPCG(2027, 42))
	anioBase := time.Now().Year() // nacimientos entre hoy-18 y hoy-7 años: siempre dentro del rango 5–20
	usados := map[string]bool{}
	nuevaPersona := func(i int) fila {
		var nombre string
		for {
			nombre = fmt.Sprintf("%s %s %s", nombres[r.IntN(len(nombres))], apellidos[r.IntN(len(apellidos))], apellidos[r.IntN(len(apellidos))])
			if !usados[nombre] {
				usados[nombre] = true
				break
			}
		}
		f := fila{
			nombre: nombre,
			numero: fmt.Sprintf("3%02d%07d", r.IntN(30)+0, 1000000+i*4999+r.IntN(4000)), // único por construcción (i crece)
			club:   clubes[r.IntN(len(clubes))], torneo: torneos[r.IntN(len(torneos))],
			pct: []int{25, 30, 40, 50, 60, 75, 100}[r.IntN(7)], autoriza: "SI",
		}
		f.correo = fmt.Sprintf("acudiente%03d@gmail.com", i)
		nac := time.Date(anioBase-18+r.IntN(12), time.Month(1+r.IntN(12)), 1+r.IntN(28), 0, 0, 0, 0, time.UTC)
		if i%10 == 0 {
			f.fechaExcel = &nac
		} else {
			f.fecha = nac.Format("02/01/2006")
		}
		return f
	}

	var filas []fila
	for i := 0; i < Validas; i++ {
		filas = append(filas, nuevaPersona(i))
	}
	idx := Validas
	// Número inválido
	malos := []string{"300123456", "2001234567", "30012345678", "abc1234567", "3-0-0", "1234567890", "300 123 45", "4001234567"}
	for i := 0; i < NumeroInvalido; i++ {
		f := nuevaPersona(idx)
		idx++
		f.numero = malos[i%len(malos)]
		filas = append(filas, f)
	}
	// Correo inválido
	correos := []string{"sin-arroba.com", "a@b", "juan@@gmail.com", "@gmail.com", "Juan <j@x.com>"}
	for i := 0; i < CorreoInvalido; i++ {
		f := nuevaPersona(idx)
		idx++
		f.correo = correos[i%len(correos)]
		filas = append(filas, f)
	}
	// Porcentaje fuera de rango
	for i, p := range []int{0, 101, 150}[:PctFuera] {
		_ = i
		f := nuevaPersona(idx)
		idx++
		f.pct = p
		filas = append(filas, f)
	}
	// Torneo inexistente
	for i := 0; i < TorneoInexistente; i++ {
		f := nuevaPersona(idx)
		idx++
		f.torneo = fmt.Sprintf("Copa Inexistente %d", 2030+i)
		filas = append(filas, f)
	}
	// autoriza = NO
	for i := 0; i < SinAutorizacion; i++ {
		f := nuevaPersona(idx)
		idx++
		f.autoriza = "NO"
		filas = append(filas, f)
	}
	// Duplicadas: copia exacta de una fila válida (mismo número, torneo y deportista)
	for i := 0; i < Duplicadas; i++ {
		d := filas[i*7]
		if d.fechaExcel != nil { // la copia usa la fecha como texto
			d.fecha = d.fechaExcel.Format("02/01/2006")
			d.fechaExcel = nil
		}
		filas = append(filas, d)
	}
	r.Shuffle(len(filas), func(i, j int) { filas[i], filas[j] = filas[j], filas[i] })

	f := excelize.NewFile()
	defer f.Close()
	f.SetSheetName("Sheet1", "becas")
	enc := []any{"nombre", "numero", "correo", "club", "fecha_nacimiento", "torneo", "porcentaje_beca", "autoriza"}
	f.SetSheetRow("becas", "A1", &enc)
	estiloFecha, _ := f.NewStyle(&excelize.Style{NumFmt: 14})
	for i, x := range filas {
		var fecha any = x.fecha
		if x.fechaExcel != nil {
			fecha = *x.fechaExcel
		}
		row := []any{x.nombre, x.numero, x.correo, x.club, fecha, x.torneo, x.pct, x.autoriza}
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		if err := f.SetSheetRow("becas", cell, &row); err != nil {
			return err
		}
		if x.fechaExcel != nil {
			c, _ := excelize.CoordinatesToCellName(5, i+2)
			f.SetCellStyle("becas", c, c, estiloFecha)
		}
	}
	f.SetColWidth("becas", "A", "H", 24)
	return f.SaveAs(ruta)
}
