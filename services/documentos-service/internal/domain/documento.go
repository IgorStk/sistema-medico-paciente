// Package domain contém as entidades do serviço de documentos (Documento e
// TipoDocumento do diagrama de classes). Não depende de banco, HTTP nem S3.
package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TipoDocumento corresponde à enumeração TipoDocumento do diagrama.
type TipoDocumento string

const (
	TipoExamePrevio TipoDocumento = "EXAME_PREVIO"
	TipoLaudo       TipoDocumento = "LAUDO"
	TipoReceita     TipoDocumento = "RECEITA"
	TipoAtestado    TipoDocumento = "ATESTADO"
)

var (
	ErrTipoInvalido        = errors.New("tipo de documento inválido")
	ErrNaoEncontrado       = errors.New("documento não encontrado")
	ErrArquivoMuitoGrande  = errors.New("arquivo maior que o limite permitido")
	ErrFormatoNaoPermitido = errors.New("formato de arquivo não permitido (use PDF, JPEG, PNG ou WEBP)")
	// ErrProntuarioIndisponivel: o prontuario-service não confirmou o anexo,
	// então o upload foi desfeito.
	ErrProntuarioIndisponivel = errors.New("não foi possível anexar o documento ao prontuário; tente novamente")
)

// ErroValidacao indica dados de entrada inválidos (vira HTTP 400).
type ErroValidacao struct{ Mensagem string }

func (e *ErroValidacao) Error() string { return e.Mensagem }

// FormatosPermitidos são os content types aceitos no upload. O tipo é
// detectado pelo conteúdo do arquivo, não pela extensão informada.
var FormatosPermitidos = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
}

// ParseTipoDocumento converte texto (sem diferenciar maiúsculas) em TipoDocumento.
func ParseTipoDocumento(s string) (TipoDocumento, error) {
	t := TipoDocumento(strings.ToUpper(strings.TrimSpace(s)))
	switch t {
	case TipoExamePrevio, TipoLaudo, TipoReceita, TipoAtestado:
		return t, nil
	}
	return "", fmt.Errorf("%w: %q", ErrTipoInvalido, s)
}

// Documento é um arquivo clínico (exame, laudo, receita ou atestado) de um paciente.
type Documento struct {
	ID           uuid.UUID     `json:"id"`
	Tipo         TipoDocumento `json:"tipo"`
	PacienteID   uuid.UUID     `json:"pacienteId"`
	ConsultaID   *uuid.UUID    `json:"consultaId,omitempty"`
	AutorID      uuid.UUID     `json:"autorId"`
	NomeOriginal string        `json:"nomeOriginal"`
	ContentType  string        `json:"contentType"`
	TamanhoBytes int64         `json:"tamanhoBytes"`
	// URI é a chave do arquivo no armazenamento (atributo "uri" do diagrama).
	// Não sai na API: o download passa sempre pelo serviço.
	URI       string    `json:"-"`
	EnviadoEm time.Time `json:"enviadoEm"`
}

// NovoDocumento são os dados informados em um upload.
type NovoDocumento struct {
	Tipo         TipoDocumento
	PacienteID   uuid.UUID
	ConsultaID   *uuid.UUID
	AutorID      uuid.UUID
	NomeOriginal string
	ContentType  string
	TamanhoBytes int64
}

// FiltroDocumentos restringe uma listagem; campos nil não filtram.
type FiltroDocumentos struct {
	PacienteID *uuid.UUID
	ConsultaID *uuid.UUID
	Tipo       *TipoDocumento
}
