package main

import (
	"log"
	"net/http"
	"os"

	"github.com/juan-camilo/becas-app/backend/internal/config"
	"github.com/juan-camilo/becas-app/backend/internal/db"
	apihttp "github.com/juan-camilo/becas-app/backend/internal/http"
	"github.com/juan-camilo/becas-app/backend/internal/repository"
	"github.com/juan-camilo/becas-app/backend/internal/service"
)

func main() {
	cfg := config.Load()
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
	srv := apihttp.Server{
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
