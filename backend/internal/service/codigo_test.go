package service

import "testing"

func TestGenerarCodigo(t *testing.T) {
	vistos := map[string]bool{}
	for i := 0; i < 1000; i++ {
		c, err := GenerarCodigo(2026)
		if err != nil {
			t.Fatal(err)
		}
		if !ReCodigo.MatchString(c) {
			t.Fatalf("formato inválido: %s", c)
		}
		vistos[c] = true
	}
	if len(vistos) < 990 { // la unicidad real la garantiza el UNIQUE con reintento
		t.Errorf("demasiadas colisiones: %d únicos de 1000", len(vistos))
	}
}
