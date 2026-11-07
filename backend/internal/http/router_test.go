package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/juan-camilo/becas-app/backend/internal/db"
	"github.com/juan-camilo/becas-app/backend/internal/repository"
	"github.com/juan-camilo/becas-app/backend/internal/service"
)

func servidor(t *testing.T) http.Handler {
	t.Helper()
	d, err := db.Abrir(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.Seed(d, "admin@horizonte.test", "Admin123*"); err != nil {
		t.Fatal(err)
	}
	u := repository.Usuarios{DB: d}
	return Server{
		Auth: service.Auth{Usuarios: u, Secret: []byte("test")}, Usuarios: u,
		Torneos: repository.Torneos{DB: d}, Becas: repository.Becas{DB: d}, CORS: "http://localhost:5173",
	}.Router()
}

func pedir(h http.Handler, metodo, ruta, cuerpo, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestRutasPublicas(t *testing.T) {
	h := servidor(t)
	if w := pedir(h, "GET", "/api/health", "", ""); w.Code != 200 {
		t.Errorf("/health: %d", w.Code)
	}
	// /verificar es público: sin token responde 404 (código inexistente), nunca 401.
	for _, ruta := range []string{"/api/verificar/BEC-2026-AAAAAA", "/api/verificar/hola"} {
		w := pedir(h, "GET", ruta, "", "")
		if w.Code != 404 || !strings.Contains(w.Body.String(), `"valido":false`) {
			t.Errorf("%s: %d %s", ruta, w.Code, w.Body)
		}
	}
}

func TestRutasProtegidasSinToken(t *testing.T) {
	h := servidor(t)
	for _, c := range [][2]string{{"GET", "/api/auth/me"}, {"GET", "/api/torneos"}, {"POST", "/api/torneos"}, {"GET", "/api/plantilla"},
		{"POST", "/api/cargas"}, {"GET", "/api/notificaciones"}, {"GET", "/api/metricas"}, {"GET", "/api/becas/1/pdf"}} {
		if w := pedir(h, c[0], c[1], "", ""); w.Code != 401 {
			t.Errorf("%s %s sin token: %d (esperaba 401)", c[0], c[1], w.Code)
		}
	}
	if w := pedir(h, "GET", "/api/torneos", "", "token-falso"); w.Code != 401 {
		t.Errorf("token falso: %d", w.Code)
	}
}

func TestLoginYAcceso(t *testing.T) {
	h := servidor(t)
	if w := pedir(h, "POST", "/api/auth/login", `{"correo":"admin@horizonte.test","password":"mala"}`, ""); w.Code != 401 {
		t.Fatalf("login malo: %d", w.Code)
	}
	if w := pedir(h, "POST", "/api/auth/login", `{no es json`, ""); w.Code != 400 {
		t.Fatalf("json inválido: %d", w.Code)
	}
	w := pedir(h, "POST", "/api/auth/login", `{"correo":"admin@horizonte.test","password":"Admin123*"}`, "")
	if w.Code != 200 {
		t.Fatalf("login bueno: %d", w.Code)
	}
	tok := strings.Split(strings.Split(w.Body.String(), `"token":"`)[1], `"`)[0]
	if w := pedir(h, "GET", "/api/torneos?activos=true", "", tok); w.Code != 200 || !strings.Contains(w.Body.String(), "Copa Horizonte 2027") {
		t.Errorf("/torneos con token: %d %s", w.Code, w.Body)
	}
	// Validaciones y errores con el formato {"error": ...}
	if w := pedir(h, "POST", "/api/torneos", `{"nombre":"Torneo X","valor_inscripcion":0,"fecha_limite_pago":"2027-01-01"}`, tok); w.Code != 400 || !strings.Contains(w.Body.String(), `"error"`) {
		t.Errorf("valor 0: %d %s", w.Code, w.Body)
	}
	if w := pedir(h, "POST", "/api/torneos", `{"nombre":"copa HORIZONTE 2027","valor_inscripcion":1000,"fecha_limite_pago":"2027-01-01"}`, tok); w.Code != 409 {
		t.Errorf("nombre duplicado: %d", w.Code)
	}
}
