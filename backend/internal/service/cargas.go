package service

import (
	"time"

	"github.com/juan-camilo/becas-app/backend/internal/domain"
	"github.com/juan-camilo/becas-app/backend/internal/repository"
)

const MaxFilas = 1000

type Cargas struct {
	Torneos repository.Torneos
	Repo    repository.Cargas
}

// Previsualizar valida y clasifica las filas y guarda la carga en estado `previsualizada`.
func (s Cargas) Previsualizar(nombreArchivo string, usuarioID int64, filas []FilaCruda) (domain.Carga, error) {
	activos, err := s.Torneos.Listar(true)
	if err != nil {
		return domain.Carga{}, err
	}
	mapa := map[string]int64{}
	for _, t := range activos {
		mapa[NormalizarTexto(t.Nombre)] = t.ID
	}
	hoy := time.Now().UTC().Truncate(24 * time.Hour)

	res := make([]Resultado, len(filas))
	for i, f := range filas {
		res[i] = ValidarFila(f, mapa, hoy)
	}
	if err := Clasificar(res, s.Repo.ExisteBeca); err != nil {
		return domain.Carga{}, err
	}
	regs := make([]repository.RegistroNuevo, len(res))
	for i, r := range res {
		regs[i] = repository.RegistroNuevo{
			Fila: r.Fila, Nombre: r.Nombre, Numero: r.Numero, Correo: r.Correo, Club: r.Club,
			FechaNac: r.FechaNac, Torneo: r.Torneo, Porcentaje: r.Porcentaje, Autoriza: r.Autoriza,
			NumeroNorm: r.NumeroNorm, TorneoID: r.TorneoID, Estado: r.Estado, Errores: r.Errores,
		}
	}
	return s.Repo.Crear(nombreArchivo, usuarioID, regs)
}
