package service

import "testing"

func resultado(fila int, nombre, numero string, torneo int64, autoriza bool) Resultado {
	return Resultado{
		FilaCruda: FilaCruda{Fila: fila}, NumeroNorm: numero, NombreNorm: NormalizarTexto(nombre),
		TorneoID: torneo, Autorizado: autoriza,
	}
}

func sinBD(string, int64, string) (bool, error) { return false, nil }

func TestDuplicadoEnArchivo(t *testing.T) {
	f := []Resultado{
		resultado(2, "Ana Pérez", "+573001112233", 1, true),
		resultado(3, "ana  perez", "+573001112233", 1, true), // mismo deportista (sin tildes/espacios)
	}
	if err := Clasificar(f, sinBD); err != nil {
		t.Fatal(err)
	}
	if f[0].Estado != "valido" || f[1].Estado != "omitido_duplicado" {
		t.Fatalf("estados: %s, %s", f[0].Estado, f[1].Estado)
	}
}

func TestDuplicadoContraBD(t *testing.T) {
	existe := func(num string, torneo int64, nombre string) (bool, error) {
		return num == "+573001112233" && torneo == 1 && nombre == "ana perez", nil
	}
	f := []Resultado{resultado(2, "Ana Pérez", "+573001112233", 1, true)}
	Clasificar(f, existe)
	if f[0].Estado != "omitido_duplicado" {
		t.Fatalf("estado: %s", f[0].Estado)
	}
}

func TestNoEsDuplicado(t *testing.T) {
	f := []Resultado{
		resultado(2, "Ana Pérez", "+573001112233", 1, true),
		resultado(3, "Ana Pérez", "+573001112233", 2, true),  // otro torneo
		resultado(4, "Luis Pérez", "+573001112233", 1, true), // hermano, mismo celular
	}
	Clasificar(f, sinBD)
	for _, r := range f {
		if r.Estado != "valido" {
			t.Errorf("fila %d debería ser válida, es %s", r.Fila, r.Estado)
		}
	}
}

func TestOrdenDeClasificacion(t *testing.T) {
	inv := resultado(2, "Ana", "+573001112233", 1, true)
	inv.Errores = []string{"x"}
	no := resultado(3, "Ana Pérez", "+573001112233", 1, false)
	ok := resultado(4, "Ana Pérez", "+573001112233", 1, true)
	f := []Resultado{inv, no, ok}
	Clasificar(f, sinBD)
	want := []string{"invalido", "omitido_sin_autorizacion", "valido"} // el NO no "reserva" la clave
	for i, w := range want {
		if f[i].Estado != w {
			t.Errorf("fila %d: %s, esperaba %s", f[i].Fila, f[i].Estado, w)
		}
	}
}
