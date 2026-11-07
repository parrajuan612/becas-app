package service

import (
	"crypto/rand"
	"fmt"
	"regexp"
)

// Alfabeto de 32 caracteres sin 0, O, 1, I (256 % 32 == 0, así que no hay sesgo al usar módulo).
const alfabeto = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

var ReCodigo = regexp.MustCompile(`^BEC-\d{4}-[A-HJ-NP-Z2-9]{6}$`)

// GenerarCodigo devuelve un código con el formato BEC-AAAA-XXXXXX.
func GenerarCodigo(anio int) (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alfabeto[int(b[i])%len(alfabeto)]
	}
	return fmt.Sprintf("BEC-%04d-%s", anio, b), nil
}
