package service

import (
	"strconv"
	"strings"
)

// Valores calcula descuento y valor a pagar con aritmética entera (redondeo half-up).
func Valores(inscripcion, porcentaje int) (descuento, aPagar int) {
	descuento = (inscripcion*porcentaje + 50) / 100
	return descuento, inscripcion - descuento
}

// FormatoCOP formatea como "$ 1.250.000".
func FormatoCOP(v int) string {
	s := strconv.Itoa(v)
	var partes []string
	for len(s) > 3 {
		partes = append([]string{s[len(s)-3:]}, partes...)
		s = s[:len(s)-3]
	}
	partes = append([]string{s}, partes...)
	return "$ " + strings.Join(partes, ".")
}
