package service

import "testing"

func TestValores(t *testing.T) {
	casos := []struct{ insc, pct, desc, pagar int }{
		{350000, 50, 175000, 175000},
		{280000, 100, 280000, 0},
		{150000, 1, 1500, 148500},
		{1001, 50, 501, 500}, // 500,5 redondea hacia arriba
		{999, 33, 330, 669},  // 329,67 -> 330
	}
	for _, c := range casos {
		d, p := Valores(c.insc, c.pct)
		if d != c.desc || p != c.pagar {
			t.Errorf("Valores(%d,%d) = %d,%d; esperaba %d,%d", c.insc, c.pct, d, p, c.desc, c.pagar)
		}
	}
}

func TestFormatoCOP(t *testing.T) {
	for in, want := range map[int]string{0: "$ 0", 999: "$ 999", 1250000: "$ 1.250.000", 350000: "$ 350.000"} {
		if got := FormatoCOP(in); got != want {
			t.Errorf("FormatoCOP(%d) = %q", in, got)
		}
	}
}
