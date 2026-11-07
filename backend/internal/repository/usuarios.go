// Package repository encapsula el acceso a SQLite.
package repository

import (
	"database/sql"

	"github.com/juan-camilo/becas-app/backend/internal/domain"
)

type Usuarios struct{ DB *sql.DB }

// PorCorreo devuelve el usuario y su hash; sql.ErrNoRows si no existe.
func (r Usuarios) PorCorreo(correo string) (domain.Usuario, string, error) {
	var u domain.Usuario
	var hash string
	err := r.DB.QueryRow(`SELECT id,correo,nombre,password_hash FROM usuarios WHERE correo=?`, correo).
		Scan(&u.ID, &u.Correo, &u.Nombre, &hash)
	return u, hash, err
}

func (r Usuarios) PorID(id int64) (domain.Usuario, error) {
	var u domain.Usuario
	err := r.DB.QueryRow(`SELECT id,correo,nombre FROM usuarios WHERE id=?`, id).
		Scan(&u.ID, &u.Correo, &u.Nombre)
	return u, err
}
