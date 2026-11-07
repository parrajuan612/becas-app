// Package http expone la API REST.
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/juan-camilo/becas-app/backend/internal/repository"
	"github.com/juan-camilo/becas-app/backend/internal/service"
)

type ctxKey struct{}

type Server struct {
	Auth       service.Auth
	Usuarios   repository.Usuarios
	Torneos    repository.Torneos
	Cargas     service.Cargas
	CargasRepo repository.Cargas
	CORS       string
}

func escribirJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func errorJSON(w http.ResponseWriter, status int, msg string) {
	escribirJSON(w, status, map[string]string{"error": msg})
}

func (s Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{s.CORS},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			escribirJSON(w, 200, map[string]string{"estado": "ok"})
		})
		r.Post("/auth/login", s.login)

		r.Group(func(r chi.Router) {
			r.Use(s.requiereAuth)
			r.Get("/auth/me", s.me)

			r.Get("/torneos", s.listarTorneos)
			r.Post("/torneos", s.crearTorneo)
			r.Put("/torneos/{id}", s.editarTorneo)
			r.Delete("/torneos/{id}", s.desactivarTorneo)

			r.Get("/plantilla", s.plantilla)

			r.Post("/cargas", s.crearCarga)
			r.Get("/cargas", s.listarCargas)
			r.Get("/cargas/{id}", s.obtenerCarga)
			r.Get("/cargas/{id}/registros", s.registrosCarga)
		})
	})
	return r
}

func (s Server) requiereAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok {
			errorJSON(w, 401, "Debe iniciar sesión")
			return
		}
		id, err := s.Auth.Verificar(tok)
		if err != nil {
			errorJSON(w, 401, "Sesión inválida o expirada")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
	})
}

func (s Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Correo   string `json:"correo"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&in); err != nil {
		errorJSON(w, 400, "Solicitud inválida")
		return
	}
	tok, u, err := s.Auth.Login(in.Correo, in.Password)
	if err == service.ErrCredenciales {
		errorJSON(w, 401, err.Error())
		return
	}
	if err != nil {
		errorJSON(w, 500, "Error interno")
		return
	}
	escribirJSON(w, 200, map[string]any{"token": tok, "usuario": u})
}

func (s Server) me(w http.ResponseWriter, r *http.Request) {
	u, err := s.Usuarios.PorID(r.Context().Value(ctxKey{}).(int64))
	if err != nil {
		errorJSON(w, 401, "Usuario no encontrado")
		return
	}
	escribirJSON(w, 200, u)
}
