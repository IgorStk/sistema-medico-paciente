package config

import (
	"slices"
	"testing"
)

func TestLoadPadroes(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("CORS_ORIGINS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Porta != "8082" || !slices.Equal(cfg.CORSOrigens, []string{"http://localhost:5173"}) {
		t.Errorf("padrões inesperados: %+v", cfg)
	}
}

func TestLoadLeAmbiente(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("CORS_ORIGINS", "http://a.test, http://b.test")

	cfg, _ := Load()
	if cfg.Porta != "9000" || !slices.Equal(cfg.CORSOrigens, []string{"http://a.test", "http://b.test"}) {
		t.Errorf("config inesperada: %+v", cfg)
	}
}
