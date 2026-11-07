package repository

import (
	"database/sql"
	"time"

	"github.com/juan-camilo/becas-app/backend/internal/domain"
)

type Notificaciones struct{ DB *sql.DB }

func (r Notificaciones) Crear(becaID int64, canal, destino, plantilla, mensaje string) error {
	_, err := r.DB.Exec(`INSERT OR IGNORE INTO notificaciones(beca_id,canal,destino,plantilla,mensaje,estado) VALUES(?,?,?,?,?,'pendiente')`,
		becaID, canal, destino, plantilla, mensaje)
	return err
}

func (r Notificaciones) ids(q string, args ...any) ([]int64, error) {
	rows, err := r.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r Notificaciones) PendientesDeCarga(cargaID int64) ([]int64, error) {
	return r.ids(`SELECT n.id FROM notificaciones n JOIN becas b ON b.id=n.beca_id WHERE b.carga_id=? AND n.estado='pendiente' ORDER BY n.id`, cargaID)
}

// FallidasDeCarga devuelve las fallidas que aún admiten reintento (< maxIntentos).
func (r Notificaciones) FallidasDeCarga(cargaID int64, maxIntentos int) ([]int64, error) {
	return r.ids(`SELECT n.id FROM notificaciones n JOIN becas b ON b.id=n.beca_id WHERE b.carga_id=? AND n.estado='fallido' AND n.intentos<? ORDER BY n.id`,
		cargaID, maxIntentos)
}

// Envio agrupa lo necesario para enviar una notificación.
type Envio struct {
	ID          int64
	Canal       string
	Destino     string
	Plantilla   string
	Mensaje     string
	Estado      string
	Intentos    int
	RutaPDF     string
	Torneo      string
	Codigo      string
	Nombre      string
	Porcentaje  int
	ValorAPagar int
	FechaLimite string
}

func (r Notificaciones) ObtenerEnvio(id int64) (Envio, error) {
	var e Envio
	err := r.DB.QueryRow(`SELECT n.id,n.canal,n.destino,n.plantilla,n.mensaje,n.estado,n.intentos,COALESCE(b.ruta_pdf,''),t.nombre,
		b.codigo,b.nombre,b.porcentaje,b.valor_a_pagar,substr(b.fecha_limite_pago,1,10)
		FROM notificaciones n JOIN becas b ON b.id=n.beca_id JOIN torneos t ON t.id=b.torneo_id WHERE n.id=?`, id).
		Scan(&e.ID, &e.Canal, &e.Destino, &e.Plantilla, &e.Mensaje, &e.Estado, &e.Intentos, &e.RutaPDF, &e.Torneo,
			&e.Codigo, &e.Nombre, &e.Porcentaje, &e.ValorAPagar, &e.FechaLimite)
	return e, err
}

// RegistrarIntento suma un intento y guarda el resultado.
func (r Notificaciones) RegistrarIntento(id int64, idExterno string, errEnvio error) error {
	now := time.Now().UTC()
	if errEnvio == nil {
		_, err := r.DB.Exec(`UPDATE notificaciones SET estado='enviado',intentos=intentos+1,ultimo_error=NULL,id_externo=?,enviado_en=?,actualizado_en=? WHERE id=?`,
			idExterno, now, now, id)
		return err
	}
	_, err := r.DB.Exec(`UPDATE notificaciones SET estado='fallido',intentos=intentos+1,ultimo_error=?,actualizado_en=? WHERE id=?`,
		errEnvio.Error(), now, id)
	return err
}

type Filtro struct {
	CargaID, TorneoID int64
	Canal, Estado     string
	Page, Size        int
}

const selItem = `SELECT n.id,n.beca_id,b.carga_id,b.nombre,t.nombre,n.canal,n.destino,n.mensaje,n.estado,n.intentos,
	COALESCE(n.ultimo_error,''),COALESCE(n.id_externo,''),n.enviado_en
	FROM notificaciones n JOIN becas b ON b.id=n.beca_id JOIN torneos t ON t.id=b.torneo_id`

func scanItem(s interface{ Scan(...any) error }) (domain.NotificacionItem, error) {
	var n domain.NotificacionItem
	err := s.Scan(&n.ID, &n.BecaID, &n.CargaID, &n.Deportista, &n.Torneo, &n.Canal, &n.Destino, &n.Mensaje, &n.Estado,
		&n.Intentos, &n.UltimoError, &n.IDExterno, &n.EnviadoEn)
	return n, err
}

func (r Notificaciones) Item(id int64) (domain.NotificacionItem, error) {
	return scanItem(r.DB.QueryRow(selItem+` WHERE n.id=?`, id))
}

func (r Notificaciones) Listar(f Filtro) ([]domain.NotificacionItem, int, error) {
	where, args := ` WHERE 1=1`, []any{}
	if f.CargaID > 0 {
		where += ` AND b.carga_id=?`
		args = append(args, f.CargaID)
	}
	if f.TorneoID > 0 {
		where += ` AND b.torneo_id=?`
		args = append(args, f.TorneoID)
	}
	if f.Canal != "" {
		where += ` AND n.canal=?`
		args = append(args, f.Canal)
	}
	if f.Estado != "" {
		where += ` AND n.estado=?`
		args = append(args, f.Estado)
	}
	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM notificaciones n JOIN becas b ON b.id=n.beca_id`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.Query(selItem+where+` ORDER BY n.id DESC LIMIT ? OFFSET ?`, append(args, f.Size, (f.Page-1)*f.Size)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.NotificacionItem{}
	for rows.Next() {
		n, err := scanItem(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, n)
	}
	return out, total, rows.Err()
}
