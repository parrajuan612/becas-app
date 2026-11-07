package service

import (
	"testing"
	"time"
)

var hoy = time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC)

func TestNormalizarNumero(t *testing.T) {
	casos := []struct {
		in   string
		want string
		ok   bool
	}{
		{"3001234567", "+573001234567", true},
		{"+57 300 123 4567", "+573001234567", true},
		{"57-300-123-4567", "+573001234567", true},
		{"(300) 123-4567", "+573001234567", true},
		{"300123456", "", false},   // 9 dígitos
		{"2001234567", "", false},  // empieza por 2
		{"30012345678", "", false}, // 11 dígitos
		{"300abc4567", "", false},
		{"", "", false},
	}
	for _, c := range casos {
		got, ok := NormalizarNumero(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("NormalizarNumero(%q) = %q,%v; esperaba %q,%v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestValidarCorreo(t *testing.T) {
	casos := map[string]bool{
		"acudiente@gmail.com": true,
		"a.b@club.com.co":     true,
		"sin-arroba.com":      false,
		"a@localhost":         false,
		"Juan <a@b.com>":      false,
		"a@b.":                false,
		"":                    false,
	}
	for in, want := range casos {
		if got := ValidarCorreo(in); got != want {
			t.Errorf("ValidarCorreo(%q) = %v; esperaba %v", in, got, want)
		}
	}
}

func TestFecha(t *testing.T) {
	casos := []struct {
		in string
		ok bool // válida y con edad 5–20
	}{
		{"14/03/2012", true},  // 14 años
		{"07/11/2021", true},  // 5 años exactos
		{"08/11/2021", false}, // 4 años
		{"07/11/2005", false}, // 21 años cumplidos hoy
		{"01/01/2030", false}, // futura
		{"41000", true},       // serial Excel = 2012-04-05
		{"texto", false},
		{"31/02/2012", false},
	}
	for _, c := range casos {
		r := ValidarFila(FilaCruda{FechaNac: c.in}, nil, hoy)
		tieneError := false
		for _, e := range r.Errores {
			if e == "Fecha de nacimiento inválida o fuera del rango de edad (5–20 años)" {
				tieneError = true
			}
		}
		if tieneError == c.ok {
			t.Errorf("fecha %q: ok esperado=%v, errores=%v", c.in, c.ok, r.Errores)
		}
	}
}

func TestPorcentaje(t *testing.T) {
	casos := []struct {
		in string
		ok bool
	}{{"50", true}, {"1", true}, {"100", true}, {"0", false}, {"101", false}, {"50%", false}, {"50.0", true}, {"50.5", false}, {"", false}}
	for _, c := range casos {
		if _, ok := ParsePorcentaje(c.in); ok != c.ok {
			t.Errorf("ParsePorcentaje(%q) ok=%v; esperaba %v", c.in, ok, c.ok)
		}
	}
}

func TestAutoriza(t *testing.T) {
	casos := []struct {
		in         string
		aut, valid bool
	}{{"SI", true, true}, {"sí", true, true}, {"Sí", true, true}, {"no", false, true}, {"NO", false, true}, {"x", false, false}, {"", false, false}}
	for _, c := range casos {
		a, v := ParseAutoriza(c.in)
		if a != c.aut || v != c.valid {
			t.Errorf("ParseAutoriza(%q) = %v,%v", c.in, a, v)
		}
	}
}

func filaOK() FilaCruda {
	return FilaCruda{Fila: 2, Nombre: "Santiago Gómez Ruiz", Numero: "3001234567", Correo: "a@gmail.com",
		Club: "Club Los Andes", FechaNac: "14/03/2012", Torneo: "  copa HORIZONTE 2027 ", Porcentaje: "50", Autoriza: "SI"}
}

func TestValidarFilaCompleta(t *testing.T) {
	torneos := map[string]int64{NormalizarTexto("Copa Horizonte 2027"): 7}
	r := ValidarFila(filaOK(), torneos, hoy)
	if len(r.Errores) != 0 || r.TorneoID != 7 || r.NumeroNorm != "+573001234567" || r.Pct != 50 {
		t.Fatalf("fila válida mal procesada: %+v", r)
	}
	// Reporta todos los errores, no solo el primero.
	mala := FilaCruda{Fila: 3}
	if n := len(ValidarFila(mala, torneos, hoy).Errores); n != 8 {
		t.Errorf("fila vacía debería tener 8 errores, tiene %d", n)
	}
	// Nombres con apóstrofe, guion y ü son válidos; con dígitos no.
	for nombre, ok := range map[string]bool{"Ana-María O'Brien": true, "Jürgen Müller": true, "Juan 3": false, "Al": false} {
		f := filaOK()
		f.Nombre = nombre
		if got := len(ValidarFila(f, torneos, hoy).Errores) == 0; got != ok {
			t.Errorf("nombre %q: válido=%v esperado=%v", nombre, got, ok)
		}
	}
}
