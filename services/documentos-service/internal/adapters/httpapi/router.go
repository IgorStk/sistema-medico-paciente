// Package httpapi é o adaptador HTTP (REST/JSON) do serviço de documentos.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Pinger é satisfeito por *pgxpool.Pool; nos testes usamos um falso.
type Pinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	servico string
	banco   Pinger
}

func NovoHandler(servico string, banco Pinger) *Handler {
	return &Handler{servico: servico, banco: banco}
}

// NovoRouter registra as rotas e aplica os middlewares.
func NovoRouter(h *Handler, origensCORS []string, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)

	// API de documentos: implementada na fase 3 (upload e download).
	mux.HandleFunc("POST /documentos", naoImplementado)
	mux.HandleFunc("GET /documentos", naoImplementado)
	mux.HandleFunc("GET /documentos/{id}", naoImplementado)
	mux.HandleFunc("GET /documentos/{id}/arquivo", naoImplementado)

	return recuperar(log, registrarRequisicao(log, cors(origensCORS, mux)))
}

type statusResposta struct {
	Status  string `json:"status"`
	Servico string `json:"servico"`
	Banco   string `json:"banco"`
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.banco.Ping(ctx); err != nil {
		escreverJSON(w, http.StatusServiceUnavailable, statusResposta{"degradado", h.servico, "indisponivel"})
		return
	}
	escreverJSON(w, http.StatusOK, statusResposta{"ok", h.servico, "ok"})
}

func naoImplementado(w http.ResponseWriter, _ *http.Request) {
	escreverErro(w, http.StatusNotImplemented, "rota ainda não implementada")
}
