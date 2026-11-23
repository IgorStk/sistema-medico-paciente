// Package httpapi é o adaptador HTTP (REST/JSON) do serviço de prontuário.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Pinger é satisfeito por *pgxpool.Pool e pelo cliente do documentos-service.
type Pinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	servico    string
	banco      Pinger
	documentos Pinger
}

func NovoHandler(servico string, banco, documentos Pinger) *Handler {
	return &Handler{servico: servico, banco: banco, documentos: documentos}
}

// NovoRouter registra as rotas e aplica os middlewares.
func NovoRouter(h *Handler, origensCORS []string, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)

	// API de prontuário: implementada nas fases 4 e 5.
	mux.HandleFunc("GET /pacientes/{pacienteId}/prontuario", naoImplementado)
	mux.HandleFunc("GET /prontuarios/{id}/historico", naoImplementado)
	mux.HandleFunc("POST /prontuarios/{id}/documentos", naoImplementado)
	mux.HandleFunc("POST /prontuarios/{id}/laudos", naoImplementado)

	return recuperar(log, registrarRequisicao(log, cors(origensCORS, mux)))
}

type statusResposta struct {
	Status       string            `json:"status"`
	Servico      string            `json:"servico"`
	Banco        string            `json:"banco"`
	Dependencias map[string]string `json:"dependencias"`
}

// health responde 503 só quando o banco próprio está fora. Se o
// documentos-service cair, o serviço continua no ar, mas avisa no corpo.
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	resp := statusResposta{
		Status:       "ok",
		Servico:      h.servico,
		Banco:        "ok",
		Dependencias: map[string]string{"documentos-service": "ok"},
	}
	if err := h.documentos.Ping(ctx); err != nil {
		resp.Dependencias["documentos-service"] = "indisponivel"
	}
	if err := h.banco.Ping(ctx); err != nil {
		resp.Status, resp.Banco = "degradado", "indisponivel"
		escreverJSON(w, http.StatusServiceUnavailable, resp)
		return
	}
	escreverJSON(w, http.StatusOK, resp)
}

func naoImplementado(w http.ResponseWriter, _ *http.Request) {
	escreverErro(w, http.StatusNotImplemented, "rota ainda não implementada")
}
