// Package domain contém as entidades do serviço de prontuário (Prontuario do
// diagrama de classes). Não depende de banco, HTTP nem de outros serviços.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProntuarioNaoEncontrado  = errors.New("prontuário não encontrado")
	ErrDocumentoNaoEncontrado   = errors.New("documento não encontrado")
	ErrDocumentoDeOutroPaciente = errors.New("documento pertence a outro paciente")
	ErrDocumentoJaAnexado       = errors.New("documento já anexado a este prontuário")
)

// Prontuario é o registro clínico de um paciente: um por paciente
// (composição 1:1 com Paciente no diagrama).
type Prontuario struct {
	ID         uuid.UUID `json:"id"`
	PacienteID uuid.UUID `json:"pacienteId"`
	CriadoEm   time.Time `json:"criadoEm"`
}

// NovoProntuario cria o prontuário de um paciente.
func NovoProntuario(pacienteID uuid.UUID, agora time.Time) Prontuario {
	return Prontuario{ID: uuid.New(), PacienteID: pacienteID, CriadoEm: agora}
}

// DocumentoAnexado liga ao prontuário um documento que vive no
// documentos-service (agregação Prontuario o-- Documento no diagrama).
type DocumentoAnexado struct {
	ProntuarioID uuid.UUID `json:"prontuarioId"`
	DocumentoID  uuid.UUID `json:"documentoId"`
	AnexadoPor   uuid.UUID `json:"anexadoPor"`
	AnexadoEm    time.Time `json:"anexadoEm"`
}

// DocumentoResumo são os metadados de um documento, como o documentos-service
// os devolve. É uma cópia local: este serviço não importa código do outro.
type DocumentoResumo struct {
	ID           uuid.UUID  `json:"id"`
	Tipo         string     `json:"tipo"`
	PacienteID   uuid.UUID  `json:"pacienteId"`
	ConsultaID   *uuid.UUID `json:"consultaId,omitempty"`
	AutorID      uuid.UUID  `json:"autorId"`
	NomeOriginal string     `json:"nomeOriginal"`
	ContentType  string     `json:"contentType"`
	TamanhoBytes int64      `json:"tamanhoBytes"`
	EnviadoEm    time.Time  `json:"enviadoEm"`
}

// ItemHistorico é uma linha do histórico: o anexo mais os metadados do documento.
type ItemHistorico struct {
	DocumentoAnexado
	Documento DocumentoResumo `json:"documento"`
}

// NovoLaudo são os dados de um laudo emitido pelo médico.
type NovoLaudo struct {
	AutorID     uuid.UUID
	ConsultaID  *uuid.UUID
	NomeArquivo string
	ContentType string
}

// EnvioDocumento é o pedido de upload enviado ao documentos-service.
type EnvioDocumento struct {
	Tipo        string
	PacienteID  uuid.UUID
	ConsultaID  *uuid.UUID
	AutorID     uuid.UUID
	NomeArquivo string
	ContentType string
}
