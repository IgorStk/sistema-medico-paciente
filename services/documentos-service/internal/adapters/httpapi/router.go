// Package httpapi é o adaptador HTTP (REST/JSON) do serviço de documentos.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"documentos-service/internal/ports"
)

// Pinger é satisfeito por *pgxpool.Pool e pelo ArmazenamentoS3.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Dependencias reúne o que os handlers precisam.
type Dependencias struct {
	Servico               string
	Log                   *slog.Logger
	Banco                 Pinger
	Armazenamento         Pinger
	Prontuario            Pinger // só informativo no /health
	Upload                ports.DocumentoUploadPort
	Download              ports.DocumentoDownloadPort
	TamanhoMaxUploadBytes int64
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

	mux.HandleFunc("POST /documentos", h.enviar)
	mux.HandleFunc("GET /documentos", h.listar)
	mux.HandleFunc("GET /documentos/{id}", h.buscar)
	mux.HandleFunc("GET /documentos/{id}/arquivo", h.baixar)

	return recuperar(h.Log, registrarRequisicao(h.Log, cors(origensCORS, mux)))
}

type statusResposta struct {
	Status        string            `json:"status"`
	Servico       string            `json:"servico"`
	Banco         string            `json:"banco"`
	Armazenamento string            `json:"armazenamento"`
	Dependencias  map[string]string `json:"dependencias"`
}

// health responde 503 se o banco ou o armazenamento estiverem fora. O
// prontuario-service aparece em "dependencias", mas não derruba o status:
// consultas e downloads continuam funcionando sem ele.
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	resp := statusResposta{Status: "ok", Servico: h.Servico, Banco: "ok", Armazenamento: "ok",
		Dependencias: map[string]string{"prontuario-service": "ok"}}
	if err := h.Prontuario.Ping(ctx); err != nil {
		resp.Dependencias["prontuario-service"] = "indisponivel"
	}
	if err := h.Banco.Ping(ctx); err != nil {
		resp.Status, resp.Banco = "degradado", "indisponivel"
	}
	if err := h.Armazenamento.Ping(ctx); err != nil {
		resp.Status, resp.Armazenamento = "degradado", "indisponivel"
	}

	status := http.StatusOK
	if resp.Status != "ok" {
		status = http.StatusServiceUnavailable
	}
	escreverJSON(w, status, resp)
}
