package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/juan-camilo/becas-app/backend/internal/db"
	"github.com/juan-camilo/becas-app/backend/internal/notifier"
	"github.com/juan-camilo/becas-app/backend/internal/repository"
)

type entorno struct {
	envio  *Envio
	cargas Cargas
	dir    string
}

func nuevoEntorno(t *testing.T, tasaFallo float64) entorno {
	t.Helper()
	d, err := db.Abrir(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.Seed(d, "admin@horizonte.test", "x"); err != nil {
		t.Fatal(err)
	}
	torneos, cargas := repository.Torneos{DB: d}, repository.Cargas{DB: d}
	dir := t.TempDir()
	envio := &Envio{
		Torneos: torneos, Cargas: cargas, Becas: repository.Becas{DB: d}, Notifs: repository.Notificaciones{DB: d},
		Notifiers: map[notifier.Canal]notifier.Notifier{
			notifier.WhatsApp: notifier.WhatsAppSimulado(tasaFallo, 0, 0),
			notifier.Correo:   notifier.CorreoSimulado(tasaFallo, 0, 0),
		},
		Workers: 5, PDFDir: dir, Sincrono: true,
	}
	return entorno{envio, Cargas{Torneos: torneos, Repo: cargas}, dir}
}

func filasEjemplo(n int) []FilaCruda {
	var out []FilaCruda
	for i := 0; i < n; i++ {
		out = append(out, FilaCruda{
			Fila: i + 2, Nombre: "Deportista Número" + string(rune('A'+i)), Numero: "30011122" + string(rune('0'+i/10)) + string(rune('0'+i%10)),
			Correo: "a@gmail.com", Club: "Club Los Andes", FechaNac: "14/03/2012", Torneo: "Copa Horizonte 2027",
			Porcentaje: "50", Autoriza: "SI",
		})
	}
	return out
}

func confirmar(t *testing.T, e entorno, filas []FilaCruda) int64 {
	t.Helper()
	c, err := e.cargas.Previsualizar("t.xlsx", 1, filas)
	if err != nil || c.Validas != len(filas) {
		t.Fatalf("previsualizar: %+v err=%v", c, err)
	}
	if err := e.envio.Confirmar(c.ID); err != nil {
		t.Fatal(err)
	}
	return c.ID
}

func TestEnvioTodoEnviado(t *testing.T) {
	e := nuevoEntorno(t, 0)
	id := confirmar(t, e, filasEjemplo(5))
	p, _ := e.cargas.Repo.Progreso(id)
	if p.Total != 10 || p.Enviadas != 10 || p.Fallidas != 0 {
		t.Fatalf("progreso: %+v", p)
	}
	c, _ := e.cargas.Repo.Obtener(id)
	if c.Estado != "completada" || c.FinProceso == nil {
		t.Errorf("carga: %+v", c)
	}
	pdfs, _ := filepath.Glob(filepath.Join(e.dir, "*.pdf"))
	if len(pdfs) != 5 {
		t.Errorf("PDFs generados: %d, esperaba 5", len(pdfs))
	}
	if b, _ := os.ReadFile(pdfs[0]); len(b) < 1000 || string(b[:4]) != "%PDF" {
		t.Error("PDF inválido")
	}
}

func TestEnvioTodoFallidoYReintentos(t *testing.T) {
	e := nuevoEntorno(t, 1)
	id := confirmar(t, e, filasEjemplo(3))
	p, _ := e.cargas.Repo.Progreso(id)
	if p.Total != 6 || p.Fallidas != 6 {
		t.Fatalf("progreso: %+v", p)
	}
	items, _, _ := e.envio.Notifs.Listar(repository.Filtro{CargaID: id, Page: 1, Size: 50})
	notif := items[0]
	if notif.Intentos != 1 || notif.UltimoError == "" {
		t.Fatalf("intentos=%d error=%q", notif.Intentos, notif.UltimoError)
	}
	// Reintento individual: 1 -> 2 -> 3, luego se rechaza.
	for want := 2; want <= 3; want++ {
		n, err := e.envio.ReintentarNotificacion(context.Background(), notif.ID)
		if err != nil || n.Intentos != want {
			t.Fatalf("reintento: intentos=%d err=%v (esperaba %d)", n.Intentos, err, want)
		}
	}
	if _, err := e.envio.ReintentarNotificacion(context.Background(), notif.ID); err != ErrMaxIntentos {
		t.Fatalf("esperaba ErrMaxIntentos, obtuve %v", err)
	}
	// Reintento en bloque: las otras 5 pasan de 1 a 3 intentos y la notificación agotada no se toca.
	for i := 0; i < 3; i++ {
		e.envio.ReintentarCarga(id)
	}
	items, _, _ = e.envio.Notifs.Listar(repository.Filtro{CargaID: id, Page: 1, Size: 50})
	for _, n := range items {
		if n.Intentos != 3 {
			t.Errorf("notificación %d con %d intentos, esperaba 3 (máximo)", n.ID, n.Intentos)
		}
	}
	if n, _ := e.envio.ReintentarCarga(id); n != 0 {
		t.Errorf("no debería quedar nada reintentable, quedan %d", n)
	}
}

func TestReintentoRecuperaFallidas(t *testing.T) {
	e := nuevoEntorno(t, 1)
	id := confirmar(t, e, filasEjemplo(2))
	for _, n := range e.envio.Notifiers { // ahora el proveedor "funciona"
		n.(*notifier.Simulado).FailureRate = 0
	}
	if n, _ := e.envio.ReintentarCarga(id); n != 4 {
		t.Fatalf("reintentadas: %d", n)
	}
	if p, _ := e.cargas.Repo.Progreso(id); p.Enviadas != 4 || p.Fallidas != 0 {
		t.Fatalf("progreso: %+v", p)
	}
}

func TestSegundaCargaEsDuplicada(t *testing.T) {
	e := nuevoEntorno(t, 0)
	filas := filasEjemplo(3)
	confirmar(t, e, filas)
	c, err := e.cargas.Previsualizar("otra.xlsx", 1, filas)
	if err != nil || c.Validas != 0 || c.OmitidasDuplicado != 3 {
		t.Fatalf("segunda carga: %+v err=%v", c, err)
	}
}

func TestConfirmarSoloUnaVez(t *testing.T) {
	e := nuevoEntorno(t, 0)
	id := confirmar(t, e, filasEjemplo(1))
	if err := e.envio.Confirmar(id); err != ErrEstadoCarga {
		t.Fatalf("esperaba ErrEstadoCarga, obtuve %v", err)
	}
}
