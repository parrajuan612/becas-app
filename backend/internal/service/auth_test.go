package service

import (
	"testing"

	"github.com/juan-camilo/becas-app/backend/internal/db"
	"github.com/juan-camilo/becas-app/backend/internal/repository"
)

func nuevoAuth(t *testing.T) Auth {
	t.Helper()
	d, err := db.Abrir(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.Seed(d, "admin@horizonte.test", "Admin123*"); err != nil {
		t.Fatal(err)
	}
	return Auth{Usuarios: repository.Usuarios{DB: d}, Secret: []byte("test")}
}

func TestLogin(t *testing.T) {
	a := nuevoAuth(t)
	tok, u, err := a.Login("Admin@Horizonte.test ", "Admin123*")
	if err != nil {
		t.Fatalf("login válido falló: %v", err)
	}
	id, err := a.Verificar(tok)
	if err != nil || id != u.ID {
		t.Fatalf("token inválido: id=%d err=%v", id, err)
	}
	if _, _, err := a.Login("admin@horizonte.test", "mala"); err != ErrCredenciales {
		t.Fatalf("esperaba ErrCredenciales, obtuve %v", err)
	}
	if _, _, err := a.Login("nadie@x.com", "x"); err != ErrCredenciales {
		t.Fatalf("esperaba ErrCredenciales, obtuve %v", err)
	}
	if _, err := a.Verificar("basura"); err == nil {
		t.Fatal("token basura aceptado")
	}
}
