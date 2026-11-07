package service

import (
	"testing"
	"time"
)

func TestMascararNombre(t *testing.T) {
	for in, want := range map[string]string{
		"Santiago Gómez Ruiz": "Santiago G. R.",
		"  ana  pérez ":       "ana P.",
		"Luis":                "Luis",
		"":                    "",
		"Ñandú Úrsula":        "Ñandú Ú.",
	} {
		if got := MascararNombre(in); got != want {
			t.Errorf("MascararNombre(%q) = %q, esperaba %q", in, got, want)
		}
	}
}

func TestRendimiento(t *testing.T) {
	ini := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	fin := ini.Add(10 * time.Second)
	ms, prom := Rendimiento(&ini, &fin, 200)
	if *ms != 10000 || *prom != 50 {
		t.Errorf("ms=%d prom=%v", *ms, *prom)
	}
	if ms, prom := Rendimiento(&ini, nil, 200); ms != nil || prom != nil {
		t.Error("sin fin no hay duración")
	}
}

func TestMetricasGlobales(t *testing.T) {
	e := nuevoEntorno(t, 1) // todo falla
	confirmar(t, e, filasEjemplo(3))
	m := Metricas{DB: e.envio.Notifs.DB}
	g, err := m.Globales()
	if err != nil {
		t.Fatal(err)
	}
	if g.TotalBecas != 3 || g.PorCanal["whatsapp"].Fallido != 3 || g.PorCanal["correo"].Fallido != 3 || g.TasaExito != 0 {
		t.Fatalf("globales: %+v", g)
	}
	if len(g.UltimasCargas) != 1 || g.UltimasCargas[0].Fallidas != 6 || g.UltimasCargas[0].DuracionMS == nil {
		t.Fatalf("últimas cargas: %+v", g.UltimasCargas)
	}
	// Tras recuperar todo, la tasa sube al 100 %.
	for _, n := range e.envio.Notifiers {
		n.(interface{ SetRate(float64) }).SetRate(0)
	}
	e.envio.ReintentarCarga(g.UltimasCargas[0].ID)
	if g, _ := m.Globales(); g.TasaExito != 100 {
		t.Errorf("tasa de éxito: %v", g.TasaExito)
	}
}
