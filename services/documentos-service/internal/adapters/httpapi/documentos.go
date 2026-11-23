package httpapi

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"documentos-service/internal/domain"
)

// folgaFormulario cobre os campos de texto e os delimitadores do multipart.
const folgaFormulario = 1 << 20

// enviar: POST /documentos (multipart/form-data)
// campos: arquivo, tipo, pacienteId, autorId, consultaId (opcional).
func (h *Handler) enviar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.TamanhoMaxUploadBytes+folgaFormulario)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var excedeu *http.MaxBytesError
		if errors.As(err, &excedeu) {
			h.responderErro(w, domain.ErrArquivoMuitoGrande)
			return
		}
		escreverErro(w, http.StatusBadRequest, "envie o formulário como multipart/form-data")
		return
	}
	defer r.MultipartForm.RemoveAll() //nolint:errcheck // limpeza de temporários

	tipo, err := domain.ParseTipoDocumento(r.FormValue("tipo"))
	if err != nil {
		h.responderErro(w, err)
		return
	}
	pacienteID, err := uuidObrigatorio(r.FormValue("pacienteId"), "pacienteId")
	if err != nil {
		h.responderErro(w, err)
		return
	}
	autorID, err := uuidObrigatorio(r.FormValue("autorId"), "autorId")
	if err != nil {
		h.responderErro(w, err)
		return
	}
	consultaID, err := uuidOpcional(r.FormValue("consultaId"), "consultaId")
	if err != nil {
		h.responderErro(w, err)
		return
	}

	arquivo, cabecalho, err := r.FormFile("arquivo")
	if err != nil {
		escreverErro(w, http.StatusBadRequest, "o campo 'arquivo' é obrigatório")
		return
	}
	defer arquivo.Close()

	contentType, err := detectarFormato(arquivo)
	if err != nil {
		h.responderErro(w, err)
		return
	}

	doc, err := h.Upload.Enviar(r.Context(), domain.NovoDocumento{
		Tipo:         tipo,
		PacienteID:   pacienteID,
		ConsultaID:   consultaID,
		AutorID:      autorID,
		NomeOriginal: cabecalho.Filename,
		ContentType:  contentType,
		TamanhoBytes: cabecalho.Size,
	}, arquivo)
	if err != nil {
		h.responderErro(w, err)
		return
	}

	w.Header().Set("Location", "/documentos/"+doc.ID.String())
	escreverJSON(w, http.StatusCreated, doc)
}

// listar: GET /documentos?pacienteId=&consultaId=&tipo=
func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var filtro domain.FiltroDocumentos
	var err error

	if filtro.PacienteID, err = uuidOpcional(q.Get("pacienteId"), "pacienteId"); err != nil {
		h.responderErro(w, err)
		return
	}
	if filtro.ConsultaID, err = uuidOpcional(q.Get("consultaId"), "consultaId"); err != nil {
		h.responderErro(w, err)
		return
	}
	if q.Get("tipo") != "" {
		tipo, err := domain.ParseTipoDocumento(q.Get("tipo"))
		if err != nil {
			h.responderErro(w, err)
			return
		}
		filtro.Tipo = &tipo
	}

	docs, err := h.Download.Listar(r.Context(), filtro)
	if err != nil {
		h.responderErro(w, err)
		return
	}
	escreverJSON(w, http.StatusOK, docs)
}

// buscar: GET /documentos/{id}
func (h *Handler) buscar(w http.ResponseWriter, r *http.Request) {
	id, err := uuidObrigatorio(r.PathValue("id"), "id")
	if err != nil {
		h.responderErro(w, err)
		return
	}
	doc, err := h.Download.BuscarPorID(r.Context(), id)
	if err != nil {
		h.responderErro(w, err)
		return
	}
	escreverJSON(w, http.StatusOK, doc)
}

// baixar: GET /documentos/{id}/arquivo[?inline=true]
// Por padrão força o download; com inline=true o navegador exibe (ex.: PDF).
func (h *Handler) baixar(w http.ResponseWriter, r *http.Request) {
	id, err := uuidObrigatorio(r.PathValue("id"), "id")
	if err != nil {
		h.responderErro(w, err)
		return
	}
	doc, conteudo, err := h.Download.Baixar(r.Context(), id)
	if err != nil {
		h.responderErro(w, err)
		return
	}
	defer conteudo.Close()

	disposicao := "attachment"
	if r.URL.Query().Get("inline") == "true" {
		disposicao = "inline"
	}
	cab := w.Header()
	cab.Set("Content-Type", doc.ContentType)
	cab.Set("Content-Length", strconv.FormatInt(doc.TamanhoBytes, 10))
	cab.Set("Content-Disposition", mime.FormatMediaType(disposicao, map[string]string{"filename": doc.NomeOriginal}))
	cab.Set("X-Content-Type-Options", "nosniff")
	cab.Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, conteudo); err != nil {
		// Os cabeçalhos já foram enviados; só resta registrar.
		h.Log.Warn("download interrompido", "documento", id, "erro", err)
	}
}

// responderErro traduz erros de domínio em status HTTP.
func (h *Handler) responderErro(w http.ResponseWriter, err error) {
	var validacao *domain.ErroValidacao
	switch {
	case errors.As(err, &validacao), errors.Is(err, domain.ErrTipoInvalido):
		escreverErro(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrNaoEncontrado):
		escreverErro(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrArquivoMuitoGrande):
		escreverErro(w, http.StatusRequestEntityTooLarge,
			fmt.Sprintf("%s (máximo %d MB)", err.Error(), h.TamanhoMaxUploadBytes>>20))
	case errors.Is(err, domain.ErrFormatoNaoPermitido):
		escreverErro(w, http.StatusUnsupportedMediaType, err.Error())
	case errors.Is(err, domain.ErrProntuarioIndisponivel):
		h.Log.Warn("upload desfeito: prontuário não confirmou o anexo", "erro", err)
		escreverErro(w, http.StatusBadGateway, domain.ErrProntuarioIndisponivel.Error())
	default:
		h.Log.Error("erro interno", "erro", err)
		escreverErro(w, http.StatusInternalServerError, "erro interno")
	}
}

// detectarFormato identifica o tipo pelo conteúdo (não confia na extensão
// nem no Content-Type enviado pelo cliente) e volta ao início do arquivo.
func detectarFormato(arquivo io.ReadSeeker) (string, error) {
	inicio := make([]byte, 512)
	n, err := io.ReadFull(arquivo, inicio)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", err
	}
	if _, err := arquivo.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	tipo, _, _ := mime.ParseMediaType(http.DetectContentType(inicio[:n]))
	return tipo, nil
}

func uuidObrigatorio(valor, campo string) (uuid.UUID, error) {
	if valor == "" {
		return uuid.Nil, &domain.ErroValidacao{Mensagem: campo + " é obrigatório"}
	}
	id, err := uuid.Parse(valor)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, &domain.ErroValidacao{Mensagem: campo + " deve ser um UUID válido"}
	}
	return id, nil
}

func uuidOpcional(valor, campo string) (*uuid.UUID, error) {
	if valor == "" {
		return nil, nil
	}
	id, err := uuidObrigatorio(valor, campo)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
