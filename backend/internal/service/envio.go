package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/juan-camilo/becas-app/backend/internal/domain"
	"github.com/juan-camilo/becas-app/backend/internal/notifier"
	"github.com/juan-camilo/becas-app/backend/internal/pdf"
	"github.com/juan-camilo/becas-app/backend/internal/repository"
)

const MaxIntentos = 3

var (
	ErrEstadoCarga    = errors.New("la carga no está en estado de previsualización")
	ErrNoReintentable = errors.New("solo se pueden reintentar notificaciones fallidas")
	ErrMaxIntentos    = errors.New("la notificación ya alcanzó el máximo de 3 intentos")
)

// Envio orquesta la confirmación de una carga: becas, PDF, notificaciones y envío con worker pool.
type Envio struct {
	Torneos   repository.Torneos
	Cargas    repository.Cargas
	Becas     repository.Becas
	Notifs    repository.Notificaciones
	Notifiers map[notifier.Canal]notifier.Notifier
	Workers   int
	PDFDir    string
	Sincrono  bool // true en pruebas: espera a que termine el procesamiento
}

// Confirmar pasa la carga a `procesando` y la procesa en segundo plano.
func (s *Envio) Confirmar(cargaID int64) error {
	ok, err := s.Cargas.IniciarProceso(cargaID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrEstadoCarga
	}
	s.lanzar(func() { s.Procesar(context.Background(), cargaID) })
	return nil
}

func (s *Envio) lanzar(f func()) {
	if s.Sincrono {
		f()
		return
	}
	go f()
}

// Reanudar retoma las cargas que quedaron en `procesando` (p. ej. tras reiniciar el servidor).
func (s *Envio) Reanudar() error {
	ids, err := s.Cargas.IDsProcesando()
	for _, id := range ids {
		id := id
		go s.Procesar(context.Background(), id)
	}
	return err
}

// Procesar es idempotente: crea las becas que falten, envía las notificaciones pendientes y cierra la carga.
func (s *Envio) Procesar(ctx context.Context, cargaID int64) {
	if err := s.crearBecas(cargaID); err != nil {
		log.Printf("carga %d: error creando becas: %v", cargaID, err)
		return
	}
	ids, err := s.Notifs.PendientesDeCarga(cargaID)
	if err != nil {
		log.Printf("carga %d: error listando pendientes: %v", cargaID, err)
		return
	}
	s.enviarTodas(ctx, ids)
	if err := s.Cargas.Completar(cargaID); err != nil {
		log.Printf("carga %d: error al completar: %v", cargaID, err)
	}
}

func (s *Envio) crearBecas(cargaID int64) error {
	regs, err := s.Becas.RegistrosSinBeca(cargaID)
	if err != nil {
		return err
	}
	for _, g := range regs {
		if err := s.crearBeca(cargaID, g); err != nil {
			return fmt.Errorf("fila %d: %w", g.Fila, err)
		}
	}
	return nil
}

func (s *Envio) crearBeca(cargaID int64, g repository.RegistroCrudo) error {
	t, err := s.Torneos.Obtener(g.TorneoID)
	if err != nil {
		return err
	}
	// Se revalida con el estado actual (el torneo pudo desactivarse entre la previsualización y la confirmación).
	mapa := map[string]int64{}
	if t.Activo {
		mapa[NormalizarTexto(t.Nombre)] = t.ID
	}
	hoy := time.Now().UTC().Truncate(24 * time.Hour)
	v := ValidarFila(FilaCruda{Fila: g.Fila, Nombre: g.Nombre, Numero: g.Numero, Correo: g.Correo, Club: g.Club,
		FechaNac: g.FechaNac, Torneo: g.Torneo, Porcentaje: g.Porcentaje, Autoriza: g.Autoriza}, mapa, hoy)
	if len(v.Errores) > 0 {
		return s.Becas.MarcarRegistro(g.ID, domain.RegistroInvalido, v.Errores)
	}

	desc, pagar := Valores(t.ValorInscripcion, v.Pct)
	limite, _ := time.Parse("2006-01-02", t.FechaLimitePago)
	nueva := repository.BecaNueva{
		RegistroID: g.ID, CargaID: cargaID, TorneoID: t.ID, Nombre: espacios(g.Nombre), NombreNorm: v.NombreNorm,
		Numero: v.NumeroNorm, Correo: g.Correo, Club: espacios(g.Club),
		FechaNacimiento: v.Fecha.Format("2006-01-02"), FechaLimitePago: t.FechaLimitePago,
		Porcentaje: v.Pct, ValorInscripcion: t.ValorInscripcion, Descuento: desc, ValorAPagar: pagar,
	}
	var becaID int64
	for intento := 0; ; intento++ {
		if nueva.Codigo, err = GenerarCodigo(time.Now().Year()); err != nil {
			return err
		}
		becaID, err = s.Becas.Crear(nueva)
		if errors.Is(err, repository.ErrCodigoUsado) && intento < 5 {
			continue
		}
		break
	}
	if errors.Is(err, repository.ErrBecaDuplicada) { // otra carga ganó la carrera
		return s.Becas.MarcarRegistro(g.ID, domain.RegistroOmitidoDuplicado, nil)
	}
	if err != nil {
		return err
	}

	// PDF
	datos := pdf.Datos{
		Codigo: nueva.Codigo, Nombre: nueva.Nombre, Club: nueva.Club, Torneo: t.Nombre,
		FechaNacimiento: v.Fecha, Edad: Edad(v.Fecha, hoy), Porcentaje: v.Pct,
		ValorInscripcion: FormatoCOP(t.ValorInscripcion), Descuento: FormatoCOP(desc), ValorAPagar: FormatoCOP(pagar),
		FechaLimite: limite, Emision: time.Now(),
	}
	ruta := filepath.Join(s.PDFDir, nueva.Codigo+".pdf")
	if b, err := pdf.Generar(datos); err != nil {
		log.Printf("beca %d: error generando PDF: %v", becaID, err)
	} else if err := os.WriteFile(ruta, b, 0o644); err != nil {
		log.Printf("beca %d: error guardando PDF: %v", becaID, err)
	} else if err := s.Becas.GuardarRutaPDF(becaID, ruta); err != nil {
		return err
	}

	// Notificaciones (WhatsApp y correo), con el mensaje ya generado
	cuerpo := Cuerpo(variables(nueva.Nombre, t.Nombre, v.Pct, pagar, limite, nueva.Codigo))
	if err := s.Notifs.Crear(becaID, string(notifier.WhatsApp), nueva.Numero, PlantillaBeca, cuerpo); err != nil {
		return err
	}
	return s.Notifs.Crear(becaID, string(notifier.Correo), nueva.Correo, PlantillaBeca, cuerpo)
}

func variables(nombre, torneo string, pct, pagar int, limite time.Time, codigo string) map[string]string {
	return map[string]string{
		"nombre": nombre, "torneo": torneo, "porcentaje": strconv.Itoa(pct),
		"valor_a_pagar": FormatoCOP(pagar), "fecha_limite": FechaCorta(limite), "codigo": codigo,
	}
}

// enviarTodas procesa las notificaciones con un pool de `Workers` goroutines.
func (s *Envio) enviarTodas(ctx context.Context, ids []int64) {
	n := s.Workers
	if n < 1 {
		n = 1
	}
	cola := make(chan int64)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range cola {
				if err := s.enviarUna(ctx, id); err != nil {
					log.Printf("notificación %d: %v", id, err)
				}
			}
		}()
	}
	for _, id := range ids {
		cola <- id
	}
	close(cola)
	wg.Wait()
}

// enviarUna hace un intento de envío y registra el resultado.
func (s *Envio) enviarUna(ctx context.Context, id int64) error {
	e, err := s.Notifs.ObtenerEnvio(id)
	if err != nil {
		return err
	}
	n, ok := s.Notifiers[notifier.Canal(e.Canal)]
	if !ok {
		return fmt.Errorf("canal desconocido: %s", e.Canal)
	}
	limite, _ := time.Parse("2006-01-02", e.FechaLimite)
	m := notifier.Mensaje{
		Destino: e.Destino, Plantilla: e.Plantilla, AdjuntoPDF: e.RutaPDF, Cuerpo: e.Mensaje,
		Variables: variables(e.Nombre, e.Torneo, e.Porcentaje, e.ValorAPagar, limite, e.Codigo),
	}
	if e.Canal == string(notifier.Correo) {
		m.Asunto = Asunto(e.Torneo)
	}
	idExt, errEnvio := n.Enviar(ctx, m)
	return s.Notifs.RegistrarIntento(id, idExt, errEnvio)
}

// ReintentarNotificacion reintenta una notificación fallida (máximo 3 intentos en total).
func (s *Envio) ReintentarNotificacion(ctx context.Context, id int64) (domain.NotificacionItem, error) {
	e, err := s.Notifs.ObtenerEnvio(id)
	if err != nil {
		return domain.NotificacionItem{}, err
	}
	if e.Estado != domain.NotifFallido {
		return domain.NotificacionItem{}, ErrNoReintentable
	}
	if e.Intentos >= MaxIntentos {
		return domain.NotificacionItem{}, ErrMaxIntentos
	}
	if err := s.enviarUna(ctx, id); err != nil {
		return domain.NotificacionItem{}, err
	}
	return s.Notifs.Item(id)
}

// ReintentarCarga reintenta en segundo plano todas las fallidas con intentos disponibles y devuelve cuántas son.
func (s *Envio) ReintentarCarga(cargaID int64) (int, error) {
	if _, err := s.Cargas.Obtener(cargaID); err != nil {
		return 0, err
	}
	ids, err := s.Notifs.FallidasDeCarga(cargaID, MaxIntentos)
	if err != nil || len(ids) == 0 {
		return 0, err
	}
	s.lanzar(func() { s.enviarTodas(context.Background(), ids) })
	return len(ids), nil
}
