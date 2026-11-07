// Package db abre SQLite, aplica migraciones y siembra datos iniciales.
package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

//go:embed migrations.sql
var migraciones string

// Abrir crea el archivo si no existe, activa WAL/FK y migra.
// Una sola conexión evita "database is locked" con el worker pool.
func Abrir(ruta string) (*sql.DB, error) {
	if ruta != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			return nil, err
		}
	}
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", ruta)
	d, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	d.SetMaxOpenConns(1)
	if _, err := d.Exec(migraciones); err != nil {
		return nil, fmt.Errorf("migraciones: %w", err)
	}
	return d, nil
}

// Seed crea el administrador y los torneos de ejemplo si las tablas están vacías.
func Seed(d *sql.DB, correo, password string) error {
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM usuarios`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		h, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			return err
		}
		if _, err := d.Exec(`INSERT INTO usuarios(correo,nombre,password_hash) VALUES(?,?,?)`,
			correo, "Administrador", string(h)); err != nil {
			return err
		}
	}
	if err := d.QueryRow(`SELECT COUNT(*) FROM torneos`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		torneos := []struct {
			nombre string
			valor  int
			fecha  string
		}{
			{"Copa Horizonte 2027", 350000, "2027-02-15"},
			{"Torneo Interclubes Sub-15", 280000, "2027-02-28"},
			{"Festival Infantil de Verano", 150000, "2027-01-31"},
		}
		for _, t := range torneos {
			if _, err := d.Exec(`INSERT INTO torneos(nombre,valor_inscripcion,fecha_limite_pago) VALUES(?,?,?)`,
				t.nombre, t.valor, t.fecha); err != nil {
				return err
			}
		}
	}
	return nil
}
