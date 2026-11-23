package config

import "testing"

func TestLoadPadroes(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DOCUMENTOS_SERVICE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Porta != "8082" || cfg.DocumentosServiceURL != "http://localhost:8081" {
		t.Errorf("padrões inesperados: %+v", cfg)
	}
}

func TestLoadRemoveBarraFinal(t *testing.T) {
	t.Setenv("DOCUMENTOS_SERVICE_URL", "http://documentos-service:8081/")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DocumentosServiceURL != "http://documentos-service:8081" {
		t.Errorf("DocumentosServiceURL = %q", cfg.DocumentosServiceURL)
	}
}

func TestLoadRejeitaURLInvalida(t *testing.T) {
	for _, v := range []string{"documentos-service:8081", "ftp://x", "http://"} {
		t.Setenv("DOCUMENTOS_SERVICE_URL", v)
		if _, err := Load(); err == nil {
			t.Errorf("esperava erro para %q", v)
		}
	}
}
