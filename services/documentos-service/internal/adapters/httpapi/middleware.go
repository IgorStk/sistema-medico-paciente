package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

// cors libera o frontend React (e outros serviços chamados pelo navegador)
// para as origens configuradas em CORS_ORIGINS. "*" libera qualquer origem.
func cors(origens []string, next http.Handler) http.Handler {
	permitidas := make(map[string]bool, len(origens))
	for _, o := range origens {
		permitidas[o] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origem := r.Header.Get("Origin")
		if origem != "" && (permitidas["*"] || permitidas[origem]) {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origem)
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			h.Set("Access-Control-Expose-Headers", "Content-Disposition")
		}
		// Preflight do navegador: responde aqui, antes do roteador.
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// respostaComStatus guarda o status HTTP para o log.
type respostaComStatus struct {
	http.ResponseWriter
	status int
}

func (r *respostaComStatus) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Unwrap permite que http.ResponseController alcance o writer original
// (necessário para Flush em downloads grandes).
func (r *respostaComStatus) Unwrap() http.ResponseWriter { return r.ResponseWriter }

func registrarRequisicao(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		rw := &respostaComStatus{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Info("requisição",
			"metodo", r.Method,
			"caminho", r.URL.Path,
			"status", rw.status,
			"duracao_ms", time.Since(inicio).Milliseconds(),
		)
	})
}

// recuperar transforma um panic em resposta 500 em vez de derrubar o serviço.
func recuperar(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler {
					panic(v)
				}
				log.Error("panic ao tratar requisição", "erro", v, "caminho", r.URL.Path)
				escreverErro(w, http.StatusInternalServerError, "erro interno")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
