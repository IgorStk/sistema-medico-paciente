package documentos

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("caminho = %q, quer /health", r.URL.Path)
		}
		w.WriteHeader(status)
	}))
	defer srv.Close()

	c := NovoCliente(srv.URL)
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping com serviço saudável: %v", err)
	}

	status = http.StatusServiceUnavailable
	if err := c.Ping(context.Background()); err == nil {
		t.Fatal("esperava erro com serviço respondendo 503")
	}

	srv.Close()
	if err := c.Ping(context.Background()); err == nil {
		t.Fatal("esperava erro com serviço fora do ar")
	}
}
