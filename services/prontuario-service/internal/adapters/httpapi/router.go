// Package httpapi é o adaptador HTTP (REST/JSON) do serviço de prontuário.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"prontuario-service/internal/ports"
)

// Pinger é satisfeito por *pgxpool.Pool; nos testes usamos um falso.
type Pinger interface {
	Ping(ctx context.Context) error
}

type Dependencias struct {
	Servico    string
	Log        *slog.Logger
	Banco      Pinger
	Prontuario ports.ProntuarioPort
}

type Handler struct {
	Dependencias
}

func NovoHandler(d Dependencias) *Handler {
	return &Handler{Dependencias: d}
}

// NovoRouter registra as rotas e aplica os middlewares.
func NovoRouter(h *Handler, origensCORS []string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)

	mux.HandleFunc("GET /pacientes/{pacienteId}/prontuario", h.obter)
	mux.HandleFunc("GET /pacientes/{pacienteId}/prontuario/historico", h.historico)
	mux.HandleFunc("POST /pacientes/{pacienteId}/prontuario/documentos", h.anexarDocumento)
	mux.HandleFunc("POST /pacientes/{pacienteId}/prontuario/laudos", h.emitirLaudo)

	return recuperar(h.Log, registrarRequisicao(h.Log, cors(origensCORS, mux)))
}

type statusResposta struct {
	Status  string `json:"status"`
	Servico string `json:"servico"`
	Banco   string `json:"banco"`
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.Banco.Ping(ctx); err != nil {
		escreverJSON(w, http.StatusServiceUnavailable, statusResposta{"degradado", h.Servico, "indisponivel"})
		return
	}
	escreverJSON(w, http.StatusOK, statusResposta{"ok", h.Servico, "ok"})
}
