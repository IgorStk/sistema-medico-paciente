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
)

type pingerFalso struct{ err error }

func (p pingerFalso) Ping(context.Context) error { return p.err }

func novoRouterTeste(errBanco error) http.Handler {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NovoRouter(NovoHandler("documentos-service", pingerFalso{errBanco}), []string{"http://localhost:5173"}, log)
}

func TestHealth(t *testing.T) {
	casos := []struct {
		nome       string
		errBanco   error
		quer       int
		querStatus string
	}{
		{"banco ok", nil, http.StatusOK, "ok"},
		{"banco fora", errors.New("conexão recusada"), http.StatusServiceUnavailable, "degradado"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			rec := httptest.NewRecorder()
			novoRouterTeste(c.errBanco).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

			if rec.Code != c.quer {
				t.Fatalf("status = %d, quer %d", rec.Code, c.quer)
			}
			var corpo statusResposta
			if err := json.NewDecoder(rec.Body).Decode(&corpo); err != nil {
				t.Fatal(err)
			}
			if corpo.Status != c.querStatus || corpo.Servico != "documentos-service" {
				t.Errorf("corpo inesperado: %+v", corpo)
			}
		})
	}
}

func TestRotasDaAPIRespondemNaoImplementado(t *testing.T) {
	r := novoRouterTeste(nil)
	for _, alvo := range []struct{ metodo, caminho string }{
		{http.MethodPost, "/documentos"},
		{http.MethodGet, "/documentos"},
		{http.MethodGet, "/documentos/123"},
		{http.MethodGet, "/documentos/123/arquivo"},
	} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(alvo.metodo, alvo.caminho, nil))
		if rec.Code != http.StatusNotImplemented {
			t.Errorf("%s %s = %d, quer 501", alvo.metodo, alvo.caminho, rec.Code)
		}
	}
}

func TestCORS(t *testing.T) {
	r := novoRouterTeste(nil)

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
