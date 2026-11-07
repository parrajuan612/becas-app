package service

import (
	"io"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard) // silencia los logs [SIMULADO]
	os.Exit(m.Run())
}
