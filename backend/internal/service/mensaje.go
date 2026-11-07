package service

import (
	"strings"
	"time"
)

const PlantillaBeca = "notificacion_beca_v1"

const cuerpoPlantilla = `Hola. La Academia Deportiva Horizonte le informa que {{nombre}}
ha obtenido una beca del {{porcentaje}}% para el torneo {{torneo}}.
Valor a pagar: {{valor_a_pagar}}. Fecha límite de pago: {{fecha_limite}}.
Código de verificación: {{codigo}}. Adjuntamos el documento oficial de la beca.`

// Cuerpo reemplaza las variables {{x}} de la plantilla notificacion_beca_v1.
func Cuerpo(vars map[string]string) string {
	s := cuerpoPlantilla
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
}

func Asunto(torneo string) string { return "Beca deportiva – " + torneo }

func FechaCorta(t time.Time) string { return t.Format("02/01/2006") }
