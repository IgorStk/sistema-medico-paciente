// Package servico implementa os casos de uso do prontuário (ProntuarioPort).
package servico

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"prontuario-service/internal/domain"
	"prontuario-service/internal/ports"
)

type Prontuario struct {
	repo  ports.ProntuarioRepository
	agora func() time.Time
}

var _ ports.ProntuarioPort = (*Prontuario)(nil)

func NovoProntuario(repo ports.ProntuarioRepository) *Prontuario {
	return &Prontuario{repo: repo, agora: time.Now}
}

// agoraUTC trunca em microssegundos, a precisão do Postgres.
func (s *Prontuario) agoraUTC() time.Time { return s.agora().UTC().Truncate(time.Microsecond) }

func (s *Prontuario) ObterOuCriar(ctx context.Context, pacienteID uuid.UUID) (domain.Prontuario, error) {
	if pacienteID == uuid.Nil {
		return domain.Prontuario{}, &domain.ErroValidacao{Mensagem: "pacienteId inválido"}
	}
	return s.repo.ObterOuCriar(ctx, domain.NovoProntuario(pacienteID, s.agoraUTC()))
}

func (s *Prontuario) AnexarDocumento(ctx context.Context, pacienteID uuid.UUID, doc domain.DocumentoAnexado) (domain.DocumentoAnexado, bool, error) {
	if err := doc.Validar(); err != nil {
		return domain.DocumentoAnexado{}, false, err
	}
	p, err := s.ObterOuCriar(ctx, pacienteID)
	if err != nil {
		return domain.DocumentoAnexado{}, false, err
	}
	doc.EnviadoEm = doc.EnviadoEm.UTC()
	doc.AnexadoEm = s.agoraUTC()
	novo, err := s.repo.AnexarDocumento(ctx, p.ID, doc)
	return doc, novo, err
}

func (s *Prontuario) EmitirLaudo(ctx context.Context, pacienteID uuid.UUID, n domain.NovoLaudo) (domain.Laudo, error) {
	if err := n.Validar(); err != nil {
		return domain.Laudo{}, err
	}
	p, err := s.ObterOuCriar(ctx, pacienteID)
	if err != nil {
		return domain.Laudo{}, err
	}
	laudo := domain.Laudo{
		ID:           uuid.New(),
		ProntuarioID: p.ID,
		ConsultaID:   n.ConsultaID,
		MedicoID:     n.MedicoID,
		Titulo:       strings.TrimSpace(n.Titulo),
		Conteudo:     strings.TrimSpace(n.Conteudo),
		EmitidoEm:    s.agoraUTC(),
	}
	if err := s.repo.SalvarLaudo(ctx, laudo); err != nil {
		return domain.Laudo{}, err
	}
	return laudo, nil
}

func (s *Prontuario) Historico(ctx context.Context, pacienteID uuid.UUID) ([]domain.ItemHistorico, error) {
	p, err := s.ObterOuCriar(ctx, pacienteID)
	if err != nil {
		return nil, err
	}
	docs, err := s.repo.ListarDocumentos(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	laudos, err := s.repo.ListarLaudos(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	return domain.MontarHistorico(docs, laudos), nil
}
