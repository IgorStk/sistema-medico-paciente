// Package ports define as interfaces (portas) da arquitetura hexagonal.
//
// Portas de entrada: os casos de uso que a API HTTP chama
// (DocumentoUploadPort e DocumentoDownloadPort do diagrama).
// Portas de saída: o que o núcleo precisa do mundo externo
// (banco de metadados e armazenamento de arquivos).
package ports

import (
	"context"
	"io"

	"github.com/google/uuid"

	"documentos-service/internal/domain"
)

// DocumentoUploadPort recebe novos documentos.
type DocumentoUploadPort interface {
	Enviar(ctx context.Context, novo domain.NovoDocumento, conteudo io.Reader) (domain.Documento, error)
}

// DocumentoDownloadPort consulta e entrega documentos. Também é a porta que o
// TriagemService (evolução futura) vai consumir.
type DocumentoDownloadPort interface {
	BuscarPorID(ctx context.Context, id uuid.UUID) (domain.Documento, error)
	Listar(ctx context.Context, filtro domain.FiltroDocumentos) ([]domain.Documento, error)
	// Baixar devolve os metadados e o conteúdo; quem chama fecha o ReadCloser.
	Baixar(ctx context.Context, id uuid.UUID) (domain.Documento, io.ReadCloser, error)
}

// DocumentoRepository guarda os metadados dos documentos (Postgres).
type DocumentoRepository interface {
	Salvar(ctx context.Context, d domain.Documento) error
	BuscarPorID(ctx context.Context, id uuid.UUID) (domain.Documento, error)
	Listar(ctx context.Context, filtro domain.FiltroDocumentos) ([]domain.Documento, error)
	Remover(ctx context.Context, id uuid.UUID) error
}

// ProntuarioGateway é a porta de saída para o prontuario-service: a seta
// "anexar" do diagrama de microsserviços. É por ela que este serviço atua
// como CONSUMIDOR (cliente HTTP em adapters/prontuario).
type ProntuarioGateway interface {
	AnexarDocumento(ctx context.Context, d domain.Documento) error
}

// ArmazenamentoArquivos guarda o conteúdo binário. O ArmazenamentoS3 do
// diagrama é o adaptador que implementa esta porta.
type ArmazenamentoArquivos interface {
	Salvar(ctx context.Context, chave string, conteudo io.Reader, tamanho int64, contentType string) error
	Abrir(ctx context.Context, chave string) (io.ReadCloser, error)
	Remover(ctx context.Context, chave string) error
}
