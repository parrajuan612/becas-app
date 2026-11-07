// Package notifier define la interfaz de envío y sus implementaciones simuladas.
// Un futuro WhatsAppMeta implementaría la misma interfaz sin tocar el resto del sistema.
package notifier

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"time"
)

type Canal string

const (
	WhatsApp Canal = "whatsapp"
	Correo   Canal = "correo"
)

type Mensaje struct {
	Destino    string            // +573001234567 o correo
	Plantilla  string            // "notificacion_beca_v1"
	Variables  map[string]string // nombre, torneo, porcentaje, valor_a_pagar, fecha_limite, codigo
	AdjuntoPDF string            // ruta del PDF
	Asunto     string            // solo correo
	Cuerpo     string            // texto final generado
}

type Notifier interface {
	Canal() Canal
	Enviar(ctx context.Context, m Mensaje) (idExterno string, err error)
}

// Simulado reproduce latencia y fallos aleatorios sin enviar nada real.
type Simulado struct {
	canal       Canal
	prefijoID   string
	FailureRate float64
	MinMS       int
	MaxMS       int
	errores     []string
}

func WhatsAppSimulado(rate float64, minMS, maxMS int) *Simulado {
	return &Simulado{canal: WhatsApp, prefijoID: "wamid.SIM", FailureRate: rate, MinMS: minMS, MaxMS: maxMS,
		errores: []string{"timeout del proveedor", "límite de tasa excedido (429)", "número no registrado en WhatsApp"}}
}

func CorreoSimulado(rate float64, minMS, maxMS int) *Simulado {
	return &Simulado{canal: Correo, prefijoID: "mail.SIM", FailureRate: rate, MinMS: minMS, MaxMS: maxMS,
		errores: []string{"timeout del proveedor", "límite de tasa excedido (429)"}}
}

func (s *Simulado) Canal() Canal { return s.canal }

func (s *Simulado) Enviar(ctx context.Context, m Mensaje) (string, error) {
	espera := s.MinMS
	if s.MaxMS > s.MinMS {
		espera += rand.IntN(s.MaxMS - s.MinMS + 1)
	}
	select {
	case <-time.After(time.Duration(espera) * time.Millisecond):
	case <-ctx.Done():
		return "", ctx.Err()
	}
	if rand.Float64() < s.FailureRate {
		msg := s.errores[rand.IntN(len(s.errores))]
		log.Printf("[SIMULADO][%s] -> %s plantilla=%s FALLO: %s", s.canal, m.Destino, m.Plantilla, msg)
		return "", errors.New(msg)
	}
	id := fmt.Sprintf("%s%012d", s.prefijoID, rand.Int64N(1_000_000_000_000))
	log.Printf("[SIMULADO][%s] -> %s plantilla=%s OK id=%s", s.canal, m.Destino, m.Plantilla, id)
	return id, nil
}

// SetRate cambia la tasa de fallo (útil en pruebas y demostraciones).
func (s *Simulado) SetRate(r float64) { s.FailureRate = r }
