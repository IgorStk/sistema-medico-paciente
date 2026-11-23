package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"prontuario-service/internal/adapters/memoria"
	"prontuario-service/internal/domain"
	"prontuario-service/internal/servico"
)

type pingerFalso struct{ err error }

func (p pingerFalso) Ping(context.Context) error { return p.err }

func novoRouterTeste(errBanco error) http.Handler {
	h := NovoHandler(Dependencias{
		Servico:    "prontuario-service",
		Log:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		Banco:      pingerFalso{errBanco},
		Prontuario: servico.NovoProntuario(memoria.NovoProntuarioRepository()),
	})
	return NovoRouter(h, []string{"http://localhost:5173"})
}

func fazer(t *testing.T, r http.Handler, metodo, caminho, corpo string) *httptest.ResponseRecorder {
	t.Helper()
	var leitor io.Reader
	if corpo != "" {
		leitor = strings.NewReader(corpo)
	}
	req := httptest.NewRequest(metodo, caminho, leitor)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	if rec := fazer(t, novoRouterTeste(nil), http.MethodGet, "/health", ""); rec.Code != http.StatusOK {
		t.Errorf("banco ok: status %d", rec.Code)
	}
	if rec := fazer(t, novoRouterTeste(errors.New("fora")), http.MethodGet, "/health", ""); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("banco fora: status %d", rec.Code)
	}
}

func TestFluxoAnexarLaudoHistorico(t *testing.T) {
	r := novoRouterTeste(nil)
	base := "/pacientes/" + uuid.NewString() + "/prontuario"

	// O prontuário nasce no primeiro acesso e é sempre o mesmo.
	var p1, p2 domain.Prontuario
	json.NewDecoder(fazer(t, r, http.MethodGet, base, "").Body).Decode(&p1)
	json.NewDecoder(fazer(t, r, http.MethodGet, base, "").Body).Decode(&p2)
	if p1.ID == uuid.Nil || p1.ID != p2.ID {
		t.Fatalf("prontuário instável: %v %v", p1.ID, p2.ID)
	}

	// Anexo vindo do documentos-service: 201 na primeira vez, 200 na repetição.
	anexo := `{"documentoId":"` + uuid.NewString() + `","tipo":"EXAME_PREVIO","nomeOriginal":"hemograma.pdf",
		"contentType":"application/pdf","tamanhoBytes":1234,"autorId":"` + uuid.NewString() + `",
		"enviadoEm":"2026-11-20T10:00:00Z"}`
	if rec := fazer(t, r, http.MethodPost, base+"/documentos", anexo); rec.Code != http.StatusCreated {
		t.Fatalf("anexar: status %d, corpo %s", rec.Code, rec.Body)
	}
	if rec := fazer(t, r, http.MethodPost, base+"/documentos", anexo); rec.Code != http.StatusOK {
		t.Fatalf("anexar de novo: status %d, quer 200", rec.Code)
	}

	// Laudo escrito pelo médico.
	laudo := `{"medicoId":"` + uuid.NewString() + `","titulo":"Laudo de hemograma","conteudo":"Sem alterações."}`
	if rec := fazer(t, r, http.MethodPost, base+"/laudos", laudo); rec.Code != http.StatusCreated {
		t.Fatalf("laudo: status %d, corpo %s", rec.Code, rec.Body)
	}

	// Histórico: laudo (mais recente) e o documento, sem duplicar.
	rec := fazer(t, r, http.MethodGet, base+"/historico", "")
	var itens []domain.ItemHistorico
	if err := json.NewDecoder(rec.Body).Decode(&itens); err != nil {
		t.Fatal(err)
	}
	if len(itens) != 2 || itens[0].Tipo != "LAUDO" || itens[1].Tipo != "DOCUMENTO" {
		t.Errorf("histórico inesperado: %+v", itens)
	}
}

func TestErros(t *testing.T) {
	r := novoRouterTeste(nil)
	base := "/pacientes/" + uuid.NewString() + "/prontuario"
	casos := []struct {
		nome, metodo, caminho, corpo string
		quer                         int
	}{
		{"paciente inválido", http.MethodGet, "/pacientes/abc/prontuario", "", http.StatusBadRequest},
		{"JSON quebrado", http.MethodPost, base + "/laudos", "{", http.StatusBadRequest},
		{"laudo sem médico", http.MethodPost, base + "/laudos", `{"titulo":"T","conteudo":"C"}`, http.StatusBadRequest},
		{"anexo tipo inválido", http.MethodPost, base + "/documentos",
			`{"documentoId":"` + uuid.NewString() + `","tipo":"FOTO","nomeOriginal":"a","autorId":"` + uuid.NewString() + `","enviadoEm":"2026-11-20T10:00:00Z"}`,
			http.StatusBadRequest},
		{"corpo grande demais", http.MethodPost, base + "/laudos", `{"conteudo":"` + string(bytes.Repeat([]byte("a"), tamanhoMaxCorpo)) + `"}`, http.StatusRequestEntityTooLarge},
	}
	for _, c := range casos {
		if rec := fazer(t, r, c.metodo, c.caminho, c.corpo); rec.Code != c.quer {
			t.Errorf("%s: status %d, quer %d (%s)", c.nome, rec.Code, c.quer, rec.Body)
		}
	}

	// Histórico vazio é [] (e não null), para o frontend não quebrar.
	if corpo := strings.TrimSpace(fazer(t, r, http.MethodGet, base+"/historico", "").Body.String()); corpo != "[]" {
		t.Errorf("histórico vazio = %s", corpo)
	}
}

func TestPreflightCORS(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/pacientes/x/prontuario/laudos", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	novoRouterTeste(nil).ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent || rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("preflight: status %d, Allow-Origin %q", rec.Code, rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
