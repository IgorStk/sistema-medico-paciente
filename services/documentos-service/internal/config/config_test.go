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

func TestLoadS3Endpoint(t *testing.T) {
	t.Setenv("S3_ENDPOINT", "https://abc.storage.supabase.co/storage/v1/s3/")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.S3.Endpoint != "https://abc.storage.supabase.co/storage/v1/s3" {
		t.Errorf("Endpoint = %q", cfg.S3.Endpoint)
	}

	t.Setenv("S3_ENDPOINT", "localhost:8333")
	if _, err := Load(); err == nil {
		t.Error("esperava erro para endpoint sem http://")
	}
}

func TestLoadProntuarioURL(t *testing.T) {
	t.Setenv("PRONTUARIO_SERVICE_URL", "")
	cfg, err := Load()
	if err != nil || cfg.ProntuarioServiceURL != "http://localhost:8082" {
		t.Errorf("padrão = %q, %v", cfg.ProntuarioServiceURL, err)
	}
	t.Setenv("PRONTUARIO_SERVICE_URL", "prontuario:8082")
	if _, err := Load(); err == nil {
		t.Error("esperava erro para URL sem http://")
	}
}
