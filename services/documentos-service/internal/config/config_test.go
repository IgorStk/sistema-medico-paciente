package config

import (
	"slices"
	"testing"
)

func TestLoadPadroes(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("MAX_UPLOAD_MB", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Porta != "8081" || cfg.TamanhoMaxUploadMB != 20 {
		t.Errorf("padrões inesperados: %+v", cfg)
	}
}

func TestLoadLeAmbiente(t *testing.T) {
	t.Setenv("PORT", "9999")
	t.Setenv("CORS_ORIGINS", "http://a.test, http://b.test ,")
	t.Setenv("MAX_UPLOAD_MB", "5")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Porta != "9999" || cfg.TamanhoMaxUploadMB != 5 {
		t.Errorf("config inesperada: %+v", cfg)
	}
	if !slices.Equal(cfg.CORSOrigens, []string{"http://a.test", "http://b.test"}) {
		t.Errorf("CORSOrigens = %v", cfg.CORSOrigens)
	}
}

func TestLoadRejeitaTamanhoInvalido(t *testing.T) {
	t.Setenv("MAX_UPLOAD_MB", "abc")
	if _, err := Load(); err == nil {
		t.Fatal("esperava erro para MAX_UPLOAD_MB inválido")
	}
}
