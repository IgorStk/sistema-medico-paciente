package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"prontuario-service/internal/domain"
)

// ProntuarioRepository implementa ports.ProntuarioRepository.
type ProntuarioRepository struct {
	pool *pgxpool.Pool
}

func NovoProntuarioRepository(pool *pgxpool.Pool) *ProntuarioRepository {
	return &ProntuarioRepository{pool: pool}
}

// ObterOuCriar usa ON CONFLICT para que duas requisições simultâneas do
// mesmo paciente não criem dois prontuários.
func (r *ProntuarioRepository) ObterOuCriar(ctx context.Context, novo domain.Prontuario) (domain.Prontuario, error) {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO prontuarios (id, paciente_id, criado_em) VALUES ($1, $2, $3)
		 ON CONFLICT (paciente_id) DO NOTHING`,
		novo.ID, novo.PacienteID, novo.CriadoEm)
	if err != nil {
		return domain.Prontuario{}, fmt.Errorf("criando prontuário: %w", err)
	}

	var p domain.Prontuario
	err = r.pool.QueryRow(ctx,
		`SELECT id, paciente_id, criado_em FROM prontuarios WHERE paciente_id = $1`, novo.PacienteID,
	).Scan(&p.ID, &p.PacienteID, &p.CriadoEm)
	if err != nil {
		return domain.Prontuario{}, fmt.Errorf("lendo prontuário: %w", err)
	}
	p.CriadoEm = p.CriadoEm.UTC()
	return p, nil
}

func (r *ProntuarioRepository) AnexarDocumento(ctx context.Context, prontuarioID uuid.UUID, d domain.DocumentoAnexado) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`INSERT INTO prontuario_documentos (prontuario_id, documento_id, tipo, nome_original,
			content_type, tamanho_bytes, consulta_id, autor_id, enviado_em, anexado_em)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (prontuario_id, documento_id) DO NOTHING`,
		prontuarioID, d.DocumentoID, d.Tipo, d.NomeOriginal, d.ContentType, d.TamanhoBytes,
		d.ConsultaID, d.AutorID, d.EnviadoEm, d.AnexadoEm)
	if err != nil {
		return false, fmt.Errorf("anexando documento: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

func (r *ProntuarioRepository) ListarDocumentos(ctx context.Context, prontuarioID uuid.UUID) ([]domain.DocumentoAnexado, error) {
	linhas, err := r.pool.Query(ctx,
		`SELECT documento_id, tipo, nome_original, content_type, tamanho_bytes, consulta_id,
			autor_id, enviado_em, anexado_em
		 FROM prontuario_documentos WHERE prontuario_id = $1 ORDER BY enviado_em DESC`, prontuarioID)
	if err != nil {
		return nil, fmt.Errorf("listando documentos: %w", err)
	}
	return pgx.CollectRows(linhas, func(l pgx.CollectableRow) (domain.DocumentoAnexado, error) {
		var d domain.DocumentoAnexado
		err := l.Scan(&d.DocumentoID, &d.Tipo, &d.NomeOriginal, &d.ContentType, &d.TamanhoBytes,
			&d.ConsultaID, &d.AutorID, &d.EnviadoEm, &d.AnexadoEm)
		d.EnviadoEm, d.AnexadoEm = d.EnviadoEm.UTC(), d.AnexadoEm.UTC()
		return d, err
	})
}

func (r *ProntuarioRepository) SalvarLaudo(ctx context.Context, l domain.Laudo) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO laudos (id, prontuario_id, consulta_id, medico_id, titulo, conteudo, emitido_em)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		l.ID, l.ProntuarioID, l.ConsultaID, l.MedicoID, l.Titulo, l.Conteudo, l.EmitidoEm)
	if err != nil {
		return fmt.Errorf("salvando laudo: %w", err)
	}
	return nil
}

func (r *ProntuarioRepository) ListarLaudos(ctx context.Context, prontuarioID uuid.UUID) ([]domain.Laudo, error) {
	linhas, err := r.pool.Query(ctx,
		`SELECT id, prontuario_id, consulta_id, medico_id, titulo, conteudo, emitido_em
		 FROM laudos WHERE prontuario_id = $1 ORDER BY emitido_em DESC`, prontuarioID)
	if err != nil {
		return nil, fmt.Errorf("listando laudos: %w", err)
	}
	return pgx.CollectRows(linhas, func(r pgx.CollectableRow) (domain.Laudo, error) {
		var l domain.Laudo
		err := r.Scan(&l.ID, &l.ProntuarioID, &l.ConsultaID, &l.MedicoID, &l.Titulo, &l.Conteudo, &l.EmitidoEm)
		l.EmitidoEm = l.EmitidoEm.UTC()
		return l, err
	})
}
