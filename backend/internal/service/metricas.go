package service

import (
	"database/sql"
	"strings"
	"time"
	"unicode"

	"github.com/juan-camilo/becas-app/backend/internal/domain"
)

type ConteoCanal struct {
	Enviado   int `json:"enviado"`
	Fallido   int `json:"fallido"`
	Pendiente int `json:"pendiente"`
}

// CargaMetricas son los conteos de una carga más su rendimiento.
type CargaMetricas struct {
	domain.Carga
	domain.Progreso
	DuracionMS    *int64   `json:"duracion_ms"`
	MSPorRegistro *float64 `json:"ms_por_registro"`
}

type MetricasGlobales struct {
	TotalBecas              int                    `json:"total_becas"`
	PorCanal                map[string]ConteoCanal `json:"por_canal"`
	OmitidasDuplicado       int                    `json:"omitidas_duplicado"`
	OmitidasSinAutorizacion int                    `json:"omitidas_sin_autorizacion"`
	TasaExito               float64                `json:"tasa_exito"` // % enviadas sobre las ya intentadas
	UltimasCargas           []CargaMetricas        `json:"ultimas_cargas"`
}

type Metricas struct{ DB *sql.DB }

func (m Metricas) Globales() (MetricasGlobales, error) {
	g := MetricasGlobales{PorCanal: map[string]ConteoCanal{"whatsapp": {}, "correo": {}}, UltimasCargas: []CargaMetricas{}}
	if err := m.DB.QueryRow(`SELECT COUNT(*) FROM becas`).Scan(&g.TotalBecas); err != nil {
		return g, err
	}
	rows, err := m.DB.Query(`SELECT canal,estado,COUNT(*) FROM notificaciones GROUP BY canal,estado`)
	if err != nil {
		return g, err
	}
	var env, fall int
	for rows.Next() {
		var canal, estado string
		var n int
		if err := rows.Scan(&canal, &estado, &n); err != nil {
			rows.Close()
			return g, err
		}
		c := g.PorCanal[canal]
		switch estado {
		case domain.NotifEnviado:
			c.Enviado, env = n, env+n
		case domain.NotifFallido:
			c.Fallido, fall = n, fall+n
		default:
			c.Pendiente = n
		}
		g.PorCanal[canal] = c
	}
	rows.Close()
	if env+fall > 0 {
		g.TasaExito = float64(env) * 100 / float64(env+fall)
	}
	// Omitidos: solo de cargas que se procesaron, para no contar previsualizaciones descartadas.
	err = m.DB.QueryRow(`SELECT COALESCE(SUM(omitidas_duplicado),0),COALESCE(SUM(omitidas_sin_autorizacion),0) FROM cargas WHERE estado IN ('procesando','completada')`).
		Scan(&g.OmitidasDuplicado, &g.OmitidasSinAutorizacion)
	if err != nil {
		return g, err
	}

	rows, err = m.DB.Query(`SELECT c.id,c.nombre_archivo,c.estado,c.total_filas,c.validas,c.invalidas,c.omitidas_duplicado,c.omitidas_sin_autorizacion,
		c.inicio_proceso,c.fin_proceso,c.creado_en,
		COUNT(n.id),COALESCE(SUM(n.estado='pendiente'),0),COALESCE(SUM(n.estado='enviado'),0),COALESCE(SUM(n.estado='fallido'),0)
		FROM cargas c LEFT JOIN becas b ON b.carga_id=c.id LEFT JOIN notificaciones n ON n.beca_id=b.id
		GROUP BY c.id ORDER BY c.id DESC LIMIT 5`)
	if err != nil {
		return g, err
	}
	defer rows.Close()
	for rows.Next() {
		var c CargaMetricas
		if err := rows.Scan(&c.ID, &c.NombreArchivo, &c.Estado, &c.TotalFilas, &c.Validas, &c.Invalidas, &c.OmitidasDuplicado,
			&c.OmitidasSinAutorizacion, &c.InicioProceso, &c.FinProceso, &c.CreadoEn,
			&c.Total, &c.Pendientes, &c.Enviadas, &c.Fallidas); err != nil {
			return g, err
		}
		c.DuracionMS, c.MSPorRegistro = Rendimiento(c.InicioProceso, c.FinProceso, c.Validas)
		g.UltimasCargas = append(g.UltimasCargas, c)
	}
	return g, rows.Err()
}

// Rendimiento calcula la duración y el promedio de ms por registro procesado.
func Rendimiento(inicio, fin *time.Time, registros int) (*int64, *float64) {
	if inicio == nil || fin == nil {
		return nil, nil
	}
	ms := fin.Sub(*inicio).Milliseconds()
	if registros <= 0 {
		return &ms, nil
	}
	prom := float64(ms) / float64(registros)
	return &ms, &prom
}

// MascararNombre deja el nombre de pila y las iniciales de los apellidos ("Santiago Gómez Ruiz" -> "Santiago G. R.").
func MascararNombre(nombre string) string {
	p := strings.Fields(nombre)
	if len(p) == 0 {
		return ""
	}
	out := []string{p[0]}
	for _, w := range p[1:] {
		out = append(out, string(unicode.ToUpper([]rune(w)[0]))+".")
	}
	return strings.Join(out, " ")
}
