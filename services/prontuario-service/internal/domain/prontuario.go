// Package domain contém as entidades do serviço de prontuário (Prontuario do
// diagrama de classes). Não depende de banco, HTTP nem de outros serviços.
package domain

import (
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrTipoDocumentoInvalido = errors.New("tipo de documento inválido")

// ErroValidacao indica dados de entrada inválidos (vira HTTP 400).
type ErroValidacao struct{ Mensagem string }

func (e *ErroValidacao) Error() string { return e.Mensagem }

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

// TiposDocumento são os mesmos da enumeração TipoDocumento do documentos-service.
var TiposDocumento = []string{"EXAME_PREVIO", "LAUDO", "RECEITA", "ATESTADO"}

// DocumentoAnexado é um documento que o documentos-service anexou ao
// prontuário (seta "anexar" do diagrama de microsserviços). Guardamos uma
// cópia dos metadados para o histórico não depender do outro serviço; o
// arquivo em si continua no documentos-service.
type DocumentoAnexado struct {
	DocumentoID  uuid.UUID  `json:"documentoId"`
	Tipo         string     `json:"tipo"`
	NomeOriginal string     `json:"nomeOriginal"`
	ContentType  string     `json:"contentType"`
	TamanhoBytes int64      `json:"tamanhoBytes"`
	ConsultaID   *uuid.UUID `json:"consultaId,omitempty"`
	AutorID      uuid.UUID  `json:"autorId"`
	EnviadoEm    time.Time  `json:"enviadoEm"`
	AnexadoEm    time.Time  `json:"anexadoEm"`
}

// Validar confere os dados recebidos do documentos-service.
func (d DocumentoAnexado) Validar() error {
	switch {
	case d.DocumentoID == uuid.Nil:
		return &ErroValidacao{Mensagem: "documentoId é obrigatório"}
	case !slices.Contains(TiposDocumento, d.Tipo):
		return ErrTipoDocumentoInvalido
	case strings.TrimSpace(d.NomeOriginal) == "":
		return &ErroValidacao{Mensagem: "nomeOriginal é obrigatório"}
	case d.AutorID == uuid.Nil:
		return &ErroValidacao{Mensagem: "autorId é obrigatório"}
	case d.EnviadoEm.IsZero():
		return &ErroValidacao{Mensagem: "enviadoEm é obrigatório"}
	}
	return nil
}

const (
	TamanhoMaxTitulo   = 200
	TamanhoMaxConteudo = 20000
)

// Laudo é o laudo escrito pelo médico diretamente no prontuário
// (operação "emitir laudo"). Laudos em PDF entram pelo documentos-service
// com tipo LAUDO e chegam aqui como DocumentoAnexado.
type Laudo struct {
	ID           uuid.UUID  `json:"id"`
	ProntuarioID uuid.UUID  `json:"prontuarioId"`
	ConsultaID   *uuid.UUID `json:"consultaId,omitempty"`
	MedicoID     uuid.UUID  `json:"medicoId"`
	Titulo       string     `json:"titulo"`
	Conteudo     string     `json:"conteudo"`
	EmitidoEm    time.Time  `json:"emitidoEm"`
}

// NovoLaudo são os dados enviados pelo médico.
type NovoLaudo struct {
	ConsultaID *uuid.UUID `json:"consultaId"`
	MedicoID   uuid.UUID  `json:"medicoId"`
	Titulo     string     `json:"titulo"`
	Conteudo   string     `json:"conteudo"`
}

func (n NovoLaudo) Validar() error {
	titulo, conteudo := strings.TrimSpace(n.Titulo), strings.TrimSpace(n.Conteudo)
	switch {
	case n.MedicoID == uuid.Nil:
		return &ErroValidacao{Mensagem: "medicoId é obrigatório"}
	case n.ConsultaID != nil && *n.ConsultaID == uuid.Nil:
		return &ErroValidacao{Mensagem: "consultaId inválido"}
	case titulo == "":
		return &ErroValidacao{Mensagem: "titulo é obrigatório"}
	case len([]rune(titulo)) > TamanhoMaxTitulo:
		return &ErroValidacao{Mensagem: "titulo muito longo"}
	case conteudo == "":
		return &ErroValidacao{Mensagem: "conteudo é obrigatório"}
	case len([]rune(conteudo)) > TamanhoMaxConteudo:
		return &ErroValidacao{Mensagem: "conteudo muito longo"}
	}
	return nil
}

// ItemHistorico é uma linha do histórico clínico: um documento anexado ou um
// laudo emitido. Exatamente um dos dois campos vem preenchido.
type ItemHistorico struct {
	Tipo      string            `json:"tipo"` // "DOCUMENTO" ou "LAUDO"
	Data      time.Time         `json:"data"`
	Documento *DocumentoAnexado `json:"documento,omitempty"`
	Laudo     *Laudo            `json:"laudo,omitempty"`
}

// MontarHistorico junta documentos e laudos, do mais recente para o mais antigo.
func MontarHistorico(docs []DocumentoAnexado, laudos []Laudo) []ItemHistorico {
	itens := make([]ItemHistorico, 0, len(docs)+len(laudos))
	for i := range docs {
		itens = append(itens, ItemHistorico{Tipo: "DOCUMENTO", Data: docs[i].EnviadoEm, Documento: &docs[i]})
	}
	for i := range laudos {
		itens = append(itens, ItemHistorico{Tipo: "LAUDO", Data: laudos[i].EmitidoEm, Laudo: &laudos[i]})
	}
	slices.SortStableFunc(itens, func(a, b ItemHistorico) int { return b.Data.Compare(a.Data) })
	return itens
}
