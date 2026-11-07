// Package config lee la configuración desde variables de entorno (.env opcional).
package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port        string
	DBPath      string
	PDFDir      string
	JWTSecret   string
	AdminCorreo string
	AdminPass   string
	FailureRate float64
	SimMinMS    int
	SimMaxMS    int
	Workers     int
	CORSOrigin  string
}

// cargarDotEnv carga KEY=VALUE de un archivo sin pisar variables ya definidas.
func cargarDotEnv(ruta string) {
	f, err := os.Open(ruta)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		k, v, ok := strings.Cut(l, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		if _, existe := os.LookupEnv(k); !existe {
			os.Setenv(k, strings.TrimSpace(v))
		}
	}
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getInt(k string, def int) int {
	if n, err := strconv.Atoi(os.Getenv(k)); err == nil {
		return n
	}
	return def
}

func getFloat(k string, def float64) float64 {
	if n, err := strconv.ParseFloat(os.Getenv(k), 64); err == nil {
		return n
	}
	return def
}

func Load() Config {
	cargarDotEnv(".env")
	cargarDotEnv("../.env")
	return Config{
		Port:        get("PORT", "8080"),
		DBPath:      get("DB_PATH", "./data/becas.db"),
		PDFDir:      get("PDF_DIR", "./data/pdfs"),
		JWTSecret:   get("JWT_SECRET", "cambia-esto-en-tu-entorno"),
		AdminCorreo: get("ADMIN_CORREO", "admin@horizonte.test"),
		AdminPass:   get("ADMIN_PASSWORD", "Admin123*"),
		FailureRate: getFloat("SIM_FAILURE_RATE", 0.05),
		SimMinMS:    getInt("SIM_MIN_MS", 50),
		SimMaxMS:    getInt("SIM_MAX_MS", 200),
		Workers:     getInt("WORKERS", 5),
		CORSOrigin:  get("CORS_ORIGIN", "http://localhost:5173"),
	}
}
