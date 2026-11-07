package repository

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/juan-camilo/becas-app/backend/internal/domain"
)

var ErrNombreDuplicado = errors.New("ya existe un torneo con ese nombre")

type Torneos struct{ DB *sql.DB }

const colsTorneo = `id,nombre,valor_inscripcion,substr(fecha_limite_pago,1,10),activo`

func scanTorneo(s interface{ Scan(...any) error }) (domain.Torneo, error) {
	var t domain.Torneo
	var act int
	err := s.Scan(&t.ID, &t.Nombre, &t.ValorInscripcion, &t.FechaLimitePago, &act)
	t.Activo = act == 1
	return t, err
}

func (r Torneos) Listar(soloActivos bool) ([]domain.Torneo, error) {
	q := `SELECT ` + colsTorneo + ` FROM torneos`
	if soloActivos {
		q += ` WHERE activo=1`
	}
	rows, err := r.DB.Query(q + ` ORDER BY nombre`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Torneo{}
	for rows.Next() {
		t, err := scanTorneo(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r Torneos) Obtener(id int64) (domain.Torneo, error) {
	return scanTorneo(r.DB.QueryRow(`SELECT `+colsTorneo+` FROM torneos WHERE id=?`, id))
}

func mapErr(err error) error {
	if err != nil && strings.Contains(err.Error(), "UNIQUE") {
		return ErrNombreDuplicado
	}
	return err
}

func (r Torneos) Crear(t domain.Torneo) (domain.Torneo, error) {
	res, err := r.DB.Exec(`INSERT INTO torneos(nombre,valor_inscripcion,fecha_limite_pago,activo) VALUES(?,?,?,1)`,
		t.Nombre, t.ValorInscripcion, t.FechaLimitePago)
	if err != nil {
		return t, mapErr(err)
	}
	id, _ := res.LastInsertId()
	return r.Obtener(id)
}

func (r Torneos) Actualizar(t domain.Torneo) (domain.Torneo, error) {
	act := 0
	if t.Activo {
		act = 1
	}
	res, err := r.DB.Exec(`UPDATE torneos SET nombre=?,valor_inscripcion=?,fecha_limite_pago=?,activo=? WHERE id=?`,
		t.Nombre, t.ValorInscripcion, t.FechaLimitePago, act, t.ID)
	if err != nil {
		return t, mapErr(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return t, sql.ErrNoRows
	}
	return r.Obtener(t.ID)
}

func (r Torneos) Desactivar(id int64) error {
	res, err := r.DB.Exec(`UPDATE torneos SET activo=0 WHERE id=?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
