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

var errFora = errors.New("fora do ar")

func novoRouterTeste(errBanco, errDocumentos error) http.Handler {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := NovoHandler("prontuario-service", pingerFalso{errBanco}, pingerFalso{errDocumentos})
	return NovoRouter(h, []string{"http://localhost:5173"}, log)
}

func TestHealth(t *testing.T) {
	casos := []struct {
		nome                  string
		errBanco, errDocs     error
		quer                  int
		querBanco, querDocSvc string
	}{
		{"tudo ok", nil, nil, http.StatusOK, "ok", "ok"},
		{"documentos fora", nil, errFora, http.StatusOK, "ok", "indisponivel"},
		{"banco fora", errFora, nil, http.StatusServiceUnavailable, "indisponivel", "ok"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			rec := httptest.NewRecorder()
			novoRouterTeste(c.errBanco, c.errDocs).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

			if rec.Code != c.quer {
				t.Fatalf("status = %d, quer %d", rec.Code, c.quer)
			}
			var corpo statusResposta
			if err := json.NewDecoder(rec.Body).Decode(&corpo); err != nil {
				t.Fatal(err)
			}
			if corpo.Banco != c.querBanco || corpo.Dependencias["documentos-service"] != c.querDocSvc {
				t.Errorf("corpo inesperado: %+v", corpo)
			}
		})
	}
}

func TestRotasDaAPIRespondemNaoImplementado(t *testing.T) {
	r := novoRouterTeste(nil, nil)
	for _, alvo := range []struct{ metodo, caminho string }{
		{http.MethodGet, "/pacientes/abc/prontuario"},
		{http.MethodGet, "/prontuarios/abc/historico"},
		{http.MethodPost, "/prontuarios/abc/documentos"},
		{http.MethodPost, "/prontuarios/abc/laudos"},
	} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(alvo.metodo, alvo.caminho, nil))
		if rec.Code != http.StatusNotImplemented {
			t.Errorf("%s %s = %d, quer 501", alvo.metodo, alvo.caminho, rec.Code)
		}
	}
}

func TestPreflightCORS(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/prontuarios/abc/laudos", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	novoRouterTeste(nil, nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent || rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("preflight: status %d, Allow-Origin %q", rec.Code, rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
