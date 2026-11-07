// Package service contiene la lógica de negocio.
package service

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/juan-camilo/becas-app/backend/internal/domain"
	"github.com/juan-camilo/becas-app/backend/internal/repository"
)

var ErrCredenciales = errors.New("correo o contraseña incorrectos")

type Auth struct {
	Usuarios repository.Usuarios
	Secret   []byte
}

// Login valida credenciales y emite un JWT de 8 horas.
func (a Auth) Login(correo, password string) (string, domain.Usuario, error) {
	u, hash, err := a.Usuarios.PorCorreo(strings.ToLower(strings.TrimSpace(correo)))
	if errors.Is(err, sql.ErrNoRows) {
		return "", u, ErrCredenciales
	}
	if err != nil {
		return "", u, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", u, ErrCredenciales
	}
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(u.ID, 10),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.Secret)
	return tok, u, err
}

// Verificar devuelve el ID de usuario de un token válido.
func (a Auth) Verificar(token string) (int64, error) {
	t, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(*jwt.Token) (any, error) {
		return a.Secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return 0, err
	}
	sub, _ := t.Claims.GetSubject()
	return strconv.ParseInt(sub, 10, 64)
}
