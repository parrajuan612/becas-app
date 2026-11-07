package main

import (
	"log"
	"net/http"
	"os"

	"github.com/juan-camilo/becas-app/backend/internal/config"
	"github.com/juan-camilo/becas-app/backend/internal/db"
	apihttp "github.com/juan-camilo/becas-app/backend/internal/http"
	"github.com/juan-camilo/becas-app/backend/internal/notifier"
	"github.com/juan-camilo/becas-app/backend/internal/repository"
	"github.com/juan-camilo/becas-app/backend/internal/service"
)

func main() {
	cfg := config.Load()
	for _, w := range cfg.Advertencias() {
		log.Printf("ADVERTENCIA: %s", w)
	}
	if err := os.MkdirAll(cfg.PDFDir, 0o755); err != nil {
		log.Fatal(err)
	}
	d, err := db.Abrir(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()
	if err := db.Seed(d, cfg.AdminCorreo, cfg.AdminPass); err != nil {
		log.Fatal(err)
	}

	usuarios := repository.Usuarios{DB: d}
	torneos := repository.Torneos{DB: d}
	cargas := repository.Cargas{DB: d}
	becas := repository.Becas{DB: d}
	notifs := repository.Notificaciones{DB: d}
	envio := &service.Envio{
		Torneos: torneos, Cargas: cargas, Becas: becas, Notifs: notifs,
		Notifiers: map[notifier.Canal]notifier.Notifier{
			notifier.WhatsApp: notifier.WhatsAppSimulado(cfg.FailureRate, cfg.SimMinMS, cfg.SimMaxMS),
			notifier.Correo:   notifier.CorreoSimulado(cfg.FailureRate, cfg.SimMinMS, cfg.SimMaxMS),
		},
		Workers: cfg.Workers, PDFDir: cfg.PDFDir,
	}
	if err := envio.Reanudar(); err != nil {
		log.Fatal(err)
	}
	srv := apihttp.Server{
		Envio: envio, Notifs: notifs, Becas: becas, Metricas: service.Metricas{DB: d},
		Auth:       service.Auth{Usuarios: usuarios, Secret: []byte(cfg.JWTSecret)},
		Usuarios:   usuarios,
		Torneos:    torneos,
		Cargas:     service.Cargas{Torneos: torneos, Repo: cargas},
		CargasRepo: cargas,
		CORS:       cfg.CORSOrigin,
	}
	log.Printf("API escuchando en :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, srv.Router()))
}
