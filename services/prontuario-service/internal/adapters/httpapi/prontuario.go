package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"prontuario-service/internal/domain"
)

const tamanhoMaxCorpo = 256 << 10 // 256 KB: sobra para um laudo de 20 mil caracteres

// obter: GET /pacientes/{pacienteId}/prontuario
func (h *Handler) obter(w http.ResponseWriter, r *http.Request) {
	pacienteID, ok := h.lerPaciente(w, r)
	if !ok {
		return
	}
	p, err := h.Prontuario.ObterOuCriar(r.Context(), pacienteID)
	if err != nil {
		h.responderErro(w, err)
		return
	}
	escreverJSON(w, http.StatusOK, p)
}

// historico: GET /pacientes/{pacienteId}/prontuario/historico
func (h *Handler) historico(w http.ResponseWriter, r *http.Request) {
	pacienteID, ok := h.lerPaciente(w, r)
	if !ok {
		return
	}
	itens, err := h.Prontuario.Historico(r.Context(), pacienteID)
	if err != nil {
		h.responderErro(w, err)
		return
	}
	escreverJSON(w, http.StatusOK, itens)
}

// anexarDocumento: POST /pacientes/{pacienteId}/prontuario/documentos
// Chamado pelo documentos-service depois de cada upload. 201 quando anexa;
// 200 quando o documento já estava anexado (pode repetir sem duplicar).
func (h *Handler) anexarDocumento(w http.ResponseWriter, r *http.Request) {
	pacienteID, ok := h.lerPaciente(w, r)
	if !ok {
		return
	}
	var doc domain.DocumentoAnexado
	if !h.lerJSON(w, r, &doc) {
		return
	}
	anexo, novo, err := h.Prontuario.AnexarDocumento(r.Context(), pacienteID, doc)
	if err != nil {
		h.responderErro(w, err)
		return
	}
	status := http.StatusOK
	if novo {
		status = http.StatusCreated
	}
	escreverJSON(w, status, anexo)
}

// emitirLaudo: POST /pacientes/{pacienteId}/prontuario/laudos
func (h *Handler) emitirLaudo(w http.ResponseWriter, r *http.Request) {
	pacienteID, ok := h.lerPaciente(w, r)
	if !ok {
		return
	}
	var novo domain.NovoLaudo
	if !h.lerJSON(w, r, &novo) {
		return
	}
	laudo, err := h.Prontuario.EmitirLaudo(r.Context(), pacienteID, novo)
	if err != nil {
		h.responderErro(w, err)
		return
	}
	escreverJSON(w, http.StatusCreated, laudo)
}

func (h *Handler) lerPaciente(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("pacienteId"))
	if err != nil || id == uuid.Nil {
		escreverErro(w, http.StatusBadRequest, "pacienteId deve ser um UUID válido")
		return uuid.Nil, false
	}
	return id, true
}

func (h *Handler) lerJSON(w http.ResponseWriter, r *http.Request, destino any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, tamanhoMaxCorpo)
	if err := json.NewDecoder(r.Body).Decode(destino); err != nil {
		var excedeu *http.MaxBytesError
		if errors.As(err, &excedeu) {
			escreverErro(w, http.StatusRequestEntityTooLarge, "corpo da requisição muito grande")
			return false
		}
		escreverErro(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return false
	}
	return true
}

func (h *Handler) responderErro(w http.ResponseWriter, err error) {
	var validacao *domain.ErroValidacao
	switch {
	case errors.As(err, &validacao), errors.Is(err, domain.ErrTipoDocumentoInvalido):
		escreverErro(w, http.StatusBadRequest, err.Error())
	default:
		h.Log.Error("erro interno", "erro", err)
		escreverErro(w, http.StatusInternalServerError, "erro interno")
	}
}
