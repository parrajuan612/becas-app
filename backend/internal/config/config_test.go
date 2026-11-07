package config

import "testing"

func TestAdvertencias(t *testing.T) {
	if n := len((Config{JWTSecret: "cambia-esto-en-tu-entorno", AdminPass: "Admin123*"}).Advertencias()); n != 2 {
		t.Errorf("esperaba 2 advertencias, hay %d", n)
	}
	if n := len((Config{JWTSecret: "otro-secreto-largo", AdminPass: "Clave-Segura-9"}).Advertencias()); n != 0 {
		t.Errorf("esperaba 0 advertencias, hay %d", n)
	}
}
