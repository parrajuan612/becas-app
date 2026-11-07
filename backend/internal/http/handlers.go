package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/juan-camilo/becas-app/backend/internal/domain"
	"github.com/juan-camilo/becas-app/backend/internal/excel"
	"github.com/juan-camilo/becas-app/backend/internal/repository"
	"github.com/juan-camilo/becas-app/backend/internal/service"
)

const maxUpload = 5 << 20 // 5 MB

func idParam(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	return id, err == nil
}

// ---------- Torneos ----------

func (s Server) listarTorneos(w http.ResponseWriter, r *http.Request) {
	ts, err := s.Torneos.Listar(r.URL.Query().Get("activos") == "true")
	if err != nil {
		errorJSON(w, 500, "Error interno")
		return
	}
	escribirJSON(w, 200, ts)
}

// leerTorneo decodifica y valida el cuerpo de crear/editar.
func leerTorneo(w http.ResponseWriter, r *http.Request) (domain.Torneo, bool) {
	var t struct {
		domain.Torneo
		Activo *bool `json:"activo"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&t); err != nil {
		errorJSON(w, 400, "Solicitud inválida")
		return t.Torneo, false
	}
	t.Nombre = strings.Join(strings.Fields(t.Nombre), " ")
	switch {
	case len(t.Nombre) < 3 || len(t.Nombre) > 100:
		errorJSON(w, 400, "El nombre debe tener entre 3 y 100 caracteres")
	case t.ValorInscripcion <= 0:
		errorJSON(w, 400, "El valor de inscripción debe ser mayor que cero")
	default:
		if _, err := time.Parse("2006-01-02", t.FechaLimitePago); err != nil {
			errorJSON(w, 400, "La fecha límite de pago es inválida (use AAAA-MM-DD)")
			return t.Torneo, false
		}
		t.Torneo.Activo = t.Activo == nil || *t.Activo
		return t.Torneo, true
	}
	return t.Torneo, false
}

func (s Server) crearTorneo(w http.ResponseWriter, r *http.Request) {
	t, ok := leerTorneo(w, r)
	if !ok {
		return
	}
	out, err := s.Torneos.Crear(t)
	if errors.Is(err, repository.ErrNombreDuplicado) {
		errorJSON(w, 409, "Ya existe un torneo con ese nombre")
		return
	}
	if err != nil {
		errorJSON(w, 500, "Error interno")
		return
	}
	escribirJSON(w, 201, out)
}

func (s Server) editarTorneo(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		errorJSON(w, 400, "Identificador inválido")
		return
	}
	t, ok := leerTorneo(w, r)
	if !ok {
		return
	}
	t.ID = id
	out, err := s.Torneos.Actualizar(t)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		errorJSON(w, 404, "Torneo no encontrado")
	case errors.Is(err, repository.ErrNombreDuplicado):
		errorJSON(w, 409, "Ya existe un torneo con ese nombre")
	case err != nil:
		errorJSON(w, 500, "Error interno")
	default:
		escribirJSON(w, 200, out)
	}
}

func (s Server) desactivarTorneo(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		errorJSON(w, 400, "Identificador inválido")
		return
	}
	err := s.Torneos.Desactivar(id)
	if errors.Is(err, sql.ErrNoRows) {
		errorJSON(w, 404, "Torneo no encontrado")
		return
	}
	if err != nil {
		errorJSON(w, 500, "Error interno")
		return
	}
	w.WriteHeader(204)
}

// ---------- Plantilla ----------

func (s Server) plantilla(w http.ResponseWriter, _ *http.Request) {
	b, err := excel.Plantilla()
	if err != nil {
		errorJSON(w, 500, "No se pudo generar la plantilla")
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="plantilla_becas.xlsx"`)
	w.Write(b)
}

// ---------- Cargas ----------

func (s Server) crearCarga(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload+1<<16)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		errorJSON(w, 413, "El archivo supera el máximo de 5 MB")
		return
	}
	file, hdr, err := r.FormFile("archivo")
	if err != nil {
		errorJSON(w, 400, "Falta el archivo (campo «archivo»)")
		return
	}
	defer file.Close()
	if !strings.HasSuffix(strings.ToLower(hdr.Filename), ".xlsx") {
		errorJSON(w, 400, "Solo se aceptan archivos .xlsx")
		return
	}
	filas, err := excel.Leer(file)
	if err != nil {
		errorJSON(w, 400, capitalizar(err.Error()))
		return
	}
	if len(filas) == 0 {
		errorJSON(w, 400, "El archivo no tiene filas con datos")
		return
	}
	if len(filas) > service.MaxFilas {
		errorJSON(w, 400, "El archivo supera el máximo de 1.000 filas")
		return
	}
	uid := r.Context().Value(ctxKey{}).(int64)
	c, err := s.Cargas.Previsualizar(hdr.Filename, uid, filas)
	if err != nil {
		errorJSON(w, 500, "Error interno")
		return
	}
	escribirJSON(w, 201, c)
}

func capitalizar(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (s Server) listarCargas(w http.ResponseWriter, _ *http.Request) {
	cs, err := s.CargasRepo.Listar()
	if err != nil {
		errorJSON(w, 500, "Error interno")
		return
	}
	escribirJSON(w, 200, cs)
}

func (s Server) obtenerCarga(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		errorJSON(w, 400, "Identificador inválido")
		return
	}
	c, err := s.CargasRepo.Obtener(id)
	if errors.Is(err, sql.ErrNoRows) {
		errorJSON(w, 404, "Carga no encontrada")
		return
	}
	if err != nil {
		errorJSON(w, 500, "Error interno")
		return
	}
	escribirJSON(w, 200, c)
}

func (s Server) registrosCarga(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		errorJSON(w, 400, "Identificador inválido")
		return
	}
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("size"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 50
	}
	regs, total, err := s.CargasRepo.Registros(id, q.Get("estado"), page, size)
	if err != nil {
		errorJSON(w, 500, "Error interno")
		return
	}
	escribirJSON(w, 200, map[string]any{"registros": regs, "total": total, "page": page, "size": size})
}
