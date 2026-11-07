package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/juan-camilo/becas-app/backend/internal/domain"
)

type Cargas struct{ DB *sql.DB }

const colsCarga = `id,nombre_archivo,estado,total_filas,validas,invalidas,omitidas_duplicado,omitidas_sin_autorizacion,inicio_proceso,fin_proceso,creado_en`

func scanCarga(s interface{ Scan(...any) error }) (domain.Carga, error) {
	var c domain.Carga
	err := s.Scan(&c.ID, &c.NombreArchivo, &c.Estado, &c.TotalFilas, &c.Validas, &c.Invalidas,
		&c.OmitidasDuplicado, &c.OmitidasSinAutorizacion, &c.InicioProceso, &c.FinProceso, &c.CreadoEn)
	return c, err
}

// RegistroNuevo es lo que se guarda por cada fila validada.
type RegistroNuevo struct {
	Fila                                                                 int
	Nombre, Numero, Correo, Club, FechaNac, Torneo, Porcentaje, Autoriza string
	NumeroNorm                                                           string
	TorneoID                                                             int64
	Estado                                                               string
	Errores                                                              []string
}

// Crear guarda la carga y sus registros en una sola transacción y calcula los conteos.
func (r Cargas) Crear(nombreArchivo string, usuarioID int64, regs []RegistroNuevo) (domain.Carga, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return domain.Carga{}, err
	}
	defer tx.Rollback()

	var c domain.Carga
	c.TotalFilas = len(regs)
	for _, g := range regs {
		switch g.Estado {
		case domain.RegistroValido:
			c.Validas++
		case domain.RegistroInvalido:
			c.Invalidas++
		case domain.RegistroOmitidoDuplicado:
			c.OmitidasDuplicado++
		case domain.RegistroOmitidoSinAutorizacion:
			c.OmitidasSinAutorizacion++
		}
	}
	res, err := tx.Exec(`INSERT INTO cargas(nombre_archivo,usuario_id,estado,total_filas,validas,invalidas,omitidas_duplicado,omitidas_sin_autorizacion)
		VALUES(?,?,?,?,?,?,?,?)`, nombreArchivo, usuarioID, domain.CargaPrevisualizada,
		c.TotalFilas, c.Validas, c.Invalidas, c.OmitidasDuplicado, c.OmitidasSinAutorizacion)
	if err != nil {
		return c, err
	}
	id, _ := res.LastInsertId()
	st, err := tx.Prepare(`INSERT INTO registros(carga_id,fila,nombre,numero,correo,club,fecha_nacimiento,torneo,porcentaje_beca,autoriza,numero_normalizado,torneo_id,estado,errores)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return c, err
	}
	defer st.Close()
	for _, g := range regs {
		errs, _ := json.Marshal(append([]string{}, g.Errores...))
		var tid any
		if g.TorneoID != 0 {
			tid = g.TorneoID
		}
		if _, err := st.Exec(id, g.Fila, g.Nombre, g.Numero, g.Correo, g.Club, g.FechaNac, g.Torneo,
			g.Porcentaje, g.Autoriza, g.NumeroNorm, tid, g.Estado, string(errs)); err != nil {
			return c, err
		}
	}
	if err := tx.Commit(); err != nil {
		return c, err
	}
	return r.Obtener(id)
}

func (r Cargas) Obtener(id int64) (domain.Carga, error) {
	return scanCarga(r.DB.QueryRow(`SELECT `+colsCarga+` FROM cargas WHERE id=?`, id))
}

func (r Cargas) Listar() ([]domain.Carga, error) {
	rows, err := r.DB.Query(`SELECT ` + colsCarga + ` FROM cargas ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Carga{}
	for rows.Next() {
		c, err := scanCarga(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Registros devuelve una página de registros y el total que cumple el filtro.
func (r Cargas) Registros(cargaID int64, estado string, page, size int) ([]domain.Registro, int, error) {
	where, args := `carga_id=?`, []any{cargaID}
	if estado != "" {
		where += ` AND estado=?`
		args = append(args, estado)
	}
	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM registros WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.Query(`SELECT id,fila,COALESCE(nombre,''),COALESCE(numero,''),COALESCE(correo,''),COALESCE(club,''),
		COALESCE(fecha_nacimiento,''),COALESCE(torneo,''),COALESCE(porcentaje_beca,''),COALESCE(autoriza,''),estado,errores
		FROM registros WHERE `+where+` ORDER BY fila LIMIT ? OFFSET ?`, append(args, size, (page-1)*size)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Registro{}
	for rows.Next() {
		var g domain.Registro
		var errs string
		if err := rows.Scan(&g.ID, &g.Fila, &g.Nombre, &g.Numero, &g.Correo, &g.Club, &g.FechaNacimiento,
			&g.Torneo, &g.PorcentajeBeca, &g.Autoriza, &g.Estado, &errs); err != nil {
			return nil, 0, err
		}
		json.Unmarshal([]byte(errs), &g.Errores)
		out = append(out, g)
	}
	return out, total, rows.Err()
}

// ExisteBeca indica si ya hay una beca con el mismo número, torneo y deportista.
func (r Cargas) ExisteBeca(numero string, torneoID int64, nombreNorm string) (bool, error) {
	var n int
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM becas WHERE numero=? AND torneo_id=? AND nombre_normalizado=?`,
		numero, torneoID, nombreNorm).Scan(&n)
	return n > 0, err
}
