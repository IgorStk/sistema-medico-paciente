// Package ports define as interfaces (portas) da arquitetura hexagonal.
package ports

import (
	"context"

	"github.com/google/uuid"

	"prontuario-service/internal/domain"
)

// ProntuarioPort reúne os casos de uso que a API HTTP chama: anexar() e
// historico() do diagrama de classes, mais a emissão de laudo. Tudo é
// endereçado pelo paciente; o prontuário é criado no primeiro acesso.
type ProntuarioPort interface {
	ObterOuCriar(ctx context.Context, pacienteID uuid.UUID) (domain.Prontuario, error)
	// AnexarDocumento é chamado pelo documentos-service depois de cada upload.
	// Anexar o mesmo documento de novo não duplica (novo == false).
	AnexarDocumento(ctx context.Context, pacienteID uuid.UUID, doc domain.DocumentoAnexado) (anexo domain.DocumentoAnexado, novo bool, err error)
	EmitirLaudo(ctx context.Context, pacienteID uuid.UUID, laudo domain.NovoLaudo) (domain.Laudo, error)
	Historico(ctx context.Context, pacienteID uuid.UUID) ([]domain.ItemHistorico, error)
}

// ProntuarioRepository guarda prontuários, anexos e laudos (prontuario_db).
type ProntuarioRepository interface {
	// ObterOuCriar devolve o prontuário do paciente; se não existir, grava `novo`.
	ObterOuCriar(ctx context.Context, novo domain.Prontuario) (domain.Prontuario, error)
	// AnexarDocumento devolve false se o documento já estava anexado.
	AnexarDocumento(ctx context.Context, prontuarioID uuid.UUID, doc domain.DocumentoAnexado) (bool, error)
	ListarDocumentos(ctx context.Context, prontuarioID uuid.UUID) ([]domain.DocumentoAnexado, error)
	SalvarLaudo(ctx context.Context, laudo domain.Laudo) error
	ListarLaudos(ctx context.Context, prontuarioID uuid.UUID) ([]domain.Laudo, error)
}
