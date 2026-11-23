package httpapi

import (
	"encoding/json"
	"net/http"
)

type erroResposta struct {
	Erro string `json:"erro"`
}

func escreverJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func escreverErro(w http.ResponseWriter, status int, mensagem string) {
	escreverJSON(w, status, erroResposta{Erro: mensagem})
}
