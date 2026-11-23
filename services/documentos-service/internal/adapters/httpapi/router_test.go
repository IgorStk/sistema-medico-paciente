package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"documentos-service/internal/adapters/memoria"
	"documentos-service/internal/servico"
)

type pingerFalso struct{ err error }

func (p pingerFalso) Ping(context.Context) error { return p.err }

const limiteTeste = 1 << 20 // 1 MB

// novoRouterTeste monta o router com o serviço real e adaptadores em memória.
func novoRouterTeste(errBanco, errArmazenamento error) http.Handler {
	svc := servico.NovoDocumentos(memoria.NovoDocumentoRepository(), memoria.NovoArmazenamento(), limiteTeste)
	h := NovoHandler(Dependencias{
		Servico:               "documentos-service",
		Log:                   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Banco:                 pingerFalso{errBanco},
		Armazenamento:         pingerFalso{errArmazenamento},
		Upload:                svc,
		Download:              svc,
		TamanhoMaxUploadBytes: limiteTeste,
	})
	return NovoRouter(h, []string{"http://localhost:5173"})
}

func TestHealth(t *testing.T) {
	fora := errors.New("fora do ar")
	casos := []struct {
		nome            string
		errBanco, errS3 error
		quer            int
		querBanco       string
		querS3          string
	}{
		{"tudo ok", nil, nil, http.StatusOK, "ok", "ok"},
		{"banco fora", fora, nil, http.StatusServiceUnavailable, "indisponivel", "ok"},
		{"s3 fora", nil, fora, http.StatusServiceUnavailable, "ok", "indisponivel"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			rec := httptest.NewRecorder()
			novoRouterTeste(c.errBanco, c.errS3).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

			if rec.Code != c.quer {
				t.Fatalf("status = %d, quer %d", rec.Code, c.quer)
			}
			var corpo statusResposta
			if err := json.NewDecoder(rec.Body).Decode(&corpo); err != nil {
				t.Fatal(err)
			}
			if corpo.Banco != c.querBanco || corpo.Armazenamento != c.querS3 {
				t.Errorf("corpo inesperado: %+v", corpo)
			}
		})
	}
}

func TestCORS(t *testing.T) {
	r := novoRouterTeste(nil, nil)

	preflight := httptest.NewRequest(http.MethodOptions, "/documentos", nil)
	preflight.Header.Set("Origin", "http://localhost:5173")
	preflight.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, preflight)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, quer 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Allow-Origin = %q", got)
	}

	estranha := httptest.NewRequest(http.MethodGet, "/health", nil)
	estranha.Header.Set("Origin", "http://malicioso.test")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, estranha)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("origem não permitida recebeu Allow-Origin = %q", got)
	}
}
