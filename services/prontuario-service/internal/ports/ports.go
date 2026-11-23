// Package ports define as interfaces (portas) da arquitetura hexagonal.
package ports

import (
	"context"
	"io"

	"github.com/google/uuid"

	"prontuario-service/internal/domain"
)

// ProntuarioPort reúne os casos de uso que a API HTTP chama
// (anexar() e historico() do diagrama, mais a emissão de laudo).
type ProntuarioPort interface {
	// ObterOuCriarPorPaciente devolve o prontuário do paciente, criando-o no primeiro acesso.
	ObterOuCriarPorPaciente(ctx context.Context, pacienteID uuid.UUID) (domain.Prontuario, error)
	Anexar(ctx context.Context, prontuarioID, documentoID, autorID uuid.UUID) (domain.DocumentoAnexado, error)
	// EmitirLaudo envia o arquivo ao documentos-service com tipo LAUDO e o anexa ao prontuário.
	EmitirLaudo(ctx context.Context, prontuarioID uuid.UUID, laudo domain.NovoLaudo, conteudo io.Reader) (domain.ItemHistorico, error)
	Historico(ctx context.Context, prontuarioID uuid.UUID) ([]domain.ItemHistorico, error)
}

// ProntuarioRepository guarda prontuários e anexos (Postgres, prontuario_db).
type ProntuarioRepository interface {
	BuscarPorID(ctx context.Context, id uuid.UUID) (domain.Prontuario, error)
	BuscarPorPaciente(ctx context.Context, pacienteID uuid.UUID) (domain.Prontuario, error)
	Criar(ctx context.Context, p domain.Prontuario) error
	Anexar(ctx context.Context, a domain.DocumentoAnexado) error
	ListarAnexos(ctx context.Context, prontuarioID uuid.UUID) ([]domain.DocumentoAnexado, error)
}

// DocumentosGateway é a porta de saída para o documentos-service. É por ela
// que este serviço atua como CONSUMIDOR (cliente HTTP em adapters/documentos).
type DocumentosGateway interface {
	BuscarDocumento(ctx context.Context, id uuid.UUID) (domain.DocumentoResumo, error)
	EnviarDocumento(ctx context.Context, envio domain.EnvioDocumento, conteudo io.Reader) (domain.DocumentoResumo, error)
}
