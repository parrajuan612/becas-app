package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/juan-camilo/becas-app/backend/internal/domain"
)

var (
	ErrBecaDuplicada = errors.New("beca duplicada")
	ErrCodigoUsado   = errors.New("código de verificación repetido")
)

type Becas struct{ DB *sql.DB }

// RegistroCrudo son los datos originales de un registro válido pendiente de procesar.
type RegistroCrudo struct {
	ID                                                                   int64
	Fila                                                                 int
	Nombre, Numero, Correo, Club, FechaNac, Torneo, Porcentaje, Autoriza string
	TorneoID                                                             int64
}

// RegistrosSinBeca lista los registros válidos de la carga que aún no tienen beca (idempotente al reanudar).
func (r Becas) RegistrosSinBeca(cargaID int64) ([]RegistroCrudo, error) {
	rows, err := r.DB.Query(`SELECT id,fila,COALESCE(nombre,''),COALESCE(numero,''),COALESCE(correo,''),COALESCE(club,''),
		COALESCE(fecha_nacimiento,''),COALESCE(torneo,''),COALESCE(porcentaje_beca,''),COALESCE(autoriza,''),COALESCE(torneo_id,0)
		FROM registros WHERE carga_id=? AND estado='valido' AND id NOT IN (SELECT registro_id FROM becas) ORDER BY fila`, cargaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RegistroCrudo
	for rows.Next() {
		var g RegistroCrudo
		if err := rows.Scan(&g.ID, &g.Fila, &g.Nombre, &g.Numero, &g.Correo, &g.Club, &g.FechaNac, &g.Torneo, &g.Porcentaje, &g.Autoriza, &g.TorneoID); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r Becas) MarcarRegistro(id int64, estado string, errores []string) error {
	e, _ := json.Marshal(append([]string{}, errores...))
	_, err := r.DB.Exec(`UPDATE registros SET estado=?, errores=? WHERE id=?`, estado, string(e), id)
	return err
}

type BecaNueva struct {
	RegistroID, CargaID, TorneoID                        int64
	Codigo, Nombre, NombreNorm, Numero, Correo, Club     string
	FechaNacimiento, FechaLimitePago                     string // AAAA-MM-DD
	Porcentaje, ValorInscripcion, Descuento, ValorAPagar int
}

// Crear inserta la beca. Devuelve ErrBecaDuplicada o ErrCodigoUsado según el UNIQUE que falle.
func (r Becas) Crear(b BecaNueva) (int64, error) {
	res, err := r.DB.Exec(`INSERT INTO becas(registro_id,carga_id,torneo_id,codigo,nombre,nombre_normalizado,numero,correo,club,
		fecha_nacimiento,porcentaje,valor_inscripcion,descuento,valor_a_pagar,fecha_limite_pago) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		b.RegistroID, b.CargaID, b.TorneoID, b.Codigo, b.Nombre, b.NombreNorm, b.Numero, b.Correo, b.Club,
		b.FechaNacimiento, b.Porcentaje, b.ValorInscripcion, b.Descuento, b.ValorAPagar, b.FechaLimitePago)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "becas.codigo"):
			return 0, ErrCodigoUsado
		case strings.Contains(err.Error(), "UNIQUE"):
			return 0, ErrBecaDuplicada
		}
		return 0, err
	}
	return res.LastInsertId()
}

func (r Becas) GuardarRutaPDF(id int64, ruta string) error {
	_, err := r.DB.Exec(`UPDATE becas SET ruta_pdf=? WHERE id=?`, ruta, id)
	return err
}

func (r Becas) Obtener(id int64) (domain.Beca, error) {
	var b domain.Beca
	var nac, lim string
	err := r.DB.QueryRow(`SELECT id,carga_id,torneo_id,codigo,nombre,numero,correo,club,substr(fecha_nacimiento,1,10),porcentaje,
		valor_inscripcion,descuento,valor_a_pagar,substr(fecha_limite_pago,1,10),COALESCE(ruta_pdf,'') FROM becas WHERE id=?`, id).
		Scan(&b.ID, &b.CargaID, &b.TorneoID, &b.Codigo, &b.Nombre, &b.Numero, &b.Correo, &b.Club, &nac, &b.Porcentaje,
			&b.ValorInscripcion, &b.Descuento, &b.ValorAPagar, &lim, &b.RutaPDF)
	if err != nil {
		return b, err
	}
	b.FechaNacimiento, _ = time.Parse("2006-01-02", nac)
	b.FechaLimitePago, _ = time.Parse("2006-01-02", lim)
	return b, nil
}

// VerificacionPublica devuelve los datos mínimos de una beca por su código.
func (r Becas) PorCodigo(codigo string) (nombre, torneo string, porcentaje int, err error) {
	err = r.DB.QueryRow(`SELECT b.nombre,t.nombre,b.porcentaje FROM becas b JOIN torneos t ON t.id=b.torneo_id WHERE b.codigo=?`, codigo).
		Scan(&nombre, &torneo, &porcentaje)
	return
}
