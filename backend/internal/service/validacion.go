package service

import (
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"

	"github.com/juan-camilo/becas-app/backend/internal/domain"
)

// FilaCruda son los valores originales de una fila del Excel.
type FilaCruda struct {
	Fila       int
	Nombre     string
	Numero     string
	Correo     string
	Club       string
	FechaNac   string
	Torneo     string
	Porcentaje string
	Autoriza   string
}

// Resultado es una fila ya validada y clasificada.
type Resultado struct {
	FilaCruda
	NumeroNorm string // E.164
	NombreNorm string // minúsculas, sin tildes (clave anti-duplicados)
	TorneoID   int64
	Fecha      time.Time
	Pct        int
	Autorizado bool
	Errores    []string
	Estado     string
}

var reNombre = regexp.MustCompile(`^\p{L}[\p{L} '\-]*$`)

func espacios(s string) string { return strings.Join(strings.Fields(s), " ") }

// NormalizarTexto pasa a minúsculas, quita tildes y espacios extra.
func NormalizarTexto(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(strings.ToLower(espacios(s))) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// NormalizarNumero devuelve el celular colombiano en formato E.164 (+573001234567).
func NormalizarNumero(s string) (string, bool) {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else if !strings.ContainsRune(" -()+.", r) {
			return "", false
		}
	}
	d := b.String()
	if len(d) == 12 && strings.HasPrefix(d, "57") {
		d = d[2:]
	}
	if len(d) != 10 || d[0] != '3' {
		return "", false
	}
	return "+57" + d, true
}

func ValidarCorreo(s string) bool {
	s = strings.TrimSpace(s)
	a, err := mail.ParseAddress(s)
	if err != nil || a.Address != s {
		return false
	}
	dom := s[strings.LastIndex(s, "@")+1:]
	return strings.Contains(dom, ".") && !strings.HasPrefix(dom, ".") && !strings.HasSuffix(dom, ".")
}

// ParseFecha acepta DD/MM/AAAA o un serial de Excel.
func ParseFecha(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if t, err := time.Parse("02/01/2006", s); err == nil {
		return t, true
	}
	if n, err := strconv.ParseFloat(s, 64); err == nil && n > 1 && n < 80000 {
		// Serial de Excel (sistema 1900): día 25569 = 1970-01-01.
		return time.Unix(0, 0).UTC().AddDate(0, 0, int(n)-25569), true
	}
	return time.Time{}, false
}

// Edad en años cumplidos a la fecha `hoy`.
func Edad(nac, hoy time.Time) int {
	e := hoy.Year() - nac.Year()
	if hoy.Month() < nac.Month() || (hoy.Month() == nac.Month() && hoy.Day() < nac.Day()) {
		e--
	}
	return e
}

// ParsePorcentaje devuelve un entero entre 1 y 100.
func ParsePorcentaje(s string) (int, bool) {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil {
		f, ferr := strconv.ParseFloat(s, 64)
		if ferr != nil || f != float64(int(f)) {
			return 0, false
		}
		n = int(f)
	}
	return n, n >= 1 && n <= 100
}

// ParseAutoriza devuelve (autoriza, válido).
func ParseAutoriza(s string) (bool, bool) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "SI", "SÍ":
		return true, true
	case "NO":
		return false, true
	}
	return false, false
}

// ValidarFila aplica las reglas de la sección 5.1 y reporta todos los errores.
// torneos: clave NormalizarTexto(nombre) -> ID de torneos activos.
func ValidarFila(f FilaCruda, torneos map[string]int64, hoy time.Time) Resultado {
	r := Resultado{FilaCruda: f}
	add := func(m string) { r.Errores = append(r.Errores, m) }

	nombre := espacios(f.Nombre)
	if n := utf8.RuneCountInString(nombre); n < 3 || n > 100 || !reNombre.MatchString(nombre) {
		add("Nombre vacío o inválido")
	}
	if n, ok := NormalizarNumero(f.Numero); ok {
		r.NumeroNorm = n
	} else {
		add("Número de celular inválido (debe tener 10 dígitos y empezar por 3)")
	}
	if !ValidarCorreo(f.Correo) {
		add("Correo electrónico inválido")
	}
	if n := utf8.RuneCountInString(espacios(f.Club)); n < 2 || n > 80 {
		add("Club vacío")
	}
	if t, ok := ParseFecha(f.FechaNac); ok && !t.After(hoy) && Edad(t, hoy) >= 5 && Edad(t, hoy) <= 20 {
		r.Fecha = t
	} else {
		add("Fecha de nacimiento inválida o fuera del rango de edad (5–20 años)")
	}
	if id, ok := torneos[NormalizarTexto(f.Torneo)]; ok {
		r.TorneoID = id
	} else {
		add("El torneo no existe o está inactivo")
	}
	if p, ok := ParsePorcentaje(f.Porcentaje); ok {
		r.Pct = p
	} else {
		add("El porcentaje debe ser un entero entre 1 y 100")
	}
	if a, ok := ParseAutoriza(f.Autoriza); ok {
		r.Autorizado = a
	} else {
		add("Autoriza debe ser SI o NO")
	}
	r.NombreNorm = NormalizarTexto(nombre)
	if len(r.Errores) > 0 {
		r.Estado = domain.RegistroInvalido
	}
	return r
}

// ExisteBeca consulta si ya hay una beca con esa clave en la base de datos.
type ExisteBeca func(numero string, torneoID int64, nombreNorm string) (bool, error)

// Clasificar asigna el estado final (sección 5.2) en este orden:
// inválido → sin autorización → duplicado en archivo → duplicado en BD → válido.
func Clasificar(filas []Resultado, existe ExisteBeca) error {
	vistos := map[string]bool{}
	for i := range filas {
		r := &filas[i]
		switch {
		case len(r.Errores) > 0:
			r.Estado = domain.RegistroInvalido
			continue
		case !r.Autorizado:
			r.Estado = domain.RegistroOmitidoSinAutorizacion
			continue
		}
		clave := r.NumeroNorm + "|" + strconv.FormatInt(r.TorneoID, 10) + "|" + r.NombreNorm
		if vistos[clave] {
			r.Estado = domain.RegistroOmitidoDuplicado
			continue
		}
		vistos[clave] = true
		dup, err := existe(r.NumeroNorm, r.TorneoID, r.NombreNorm)
		if err != nil {
			return err
		}
		if dup {
			r.Estado = domain.RegistroOmitidoDuplicado
		} else {
			r.Estado = domain.RegistroValido
		}
	}
	return nil
}
