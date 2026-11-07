package ejemplo

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/juan-camilo/becas-app/backend/internal/excel"
	"github.com/juan-camilo/becas-app/backend/internal/service"
)

func clasificar(t *testing.T, ruta string) map[string]int {
	t.Helper()
	f, err := os.Open(ruta)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	filas, err := excel.Leer(f)
	if err != nil {
		t.Fatal(err)
	}
	torneos := map[string]int64{}
	for i, n := range torneos2 {
		torneos[service.NormalizarTexto(n)] = int64(i + 1)
	}
	hoy := time.Now().UTC().Truncate(24 * time.Hour)
	res := make([]service.Resultado, len(filas))
	for i, fc := range filas {
		res[i] = service.ValidarFila(fc, torneos, hoy)
	}
	if err := service.Clasificar(res, func(string, int64, string) (bool, error) { return false, nil }); err != nil {
		t.Fatal(err)
	}
	conteo := map[string]int{"total": len(res)}
	for _, r := range res {
		conteo[r.Estado]++
	}
	return conteo
}

var torneos2 = torneos

func TestComposicionDelEjemplo(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "becas_200.xlsx")
	if err := Generar(ruta); err != nil {
		t.Fatal(err)
	}
	c := clasificar(t, ruta)
	if c["total"] != 200 || c["valido"] != 170 || c["invalido"] != 20 || c["omitido_duplicado"] != 5 || c["omitido_sin_autorizacion"] != 5 {
		t.Fatalf("conteos inesperados: %v", c)
	}
}

func TestEjemploEsReproducible(t *testing.T) {
	d := t.TempDir()
	a, b := filepath.Join(d, "a.xlsx"), filepath.Join(d, "b.xlsx")
	Generar(a)
	Generar(b)
	fa, _ := excel2rows(a)
	fb, _ := excel2rows(b)
	if len(fa) != len(fb) {
		t.Fatal("distinto tamaño")
	}
	for i := range fa {
		if fa[i] != fb[i] {
			t.Fatalf("fila %d distinta: %+v vs %+v", i, fa[i], fb[i])
		}
	}
}

func excel2rows(ruta string) ([]service.FilaCruda, error) {
	f, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return excel.Leer(f)
}
