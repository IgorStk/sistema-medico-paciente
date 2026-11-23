package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"documentos-service/internal/domain"
)

// limiteListagem evita respostas gigantes enquanto não há paginação.
const limiteListagem = 500

const colunasDocumento = `id, tipo, paciente_id, consulta_id, autor_id, nome_original,
	content_type, tamanho_bytes, chave_armazenamento, enviado_em`

// DocumentoRepository implementa ports.DocumentoRepository.
type DocumentoRepository struct {
	pool *pgxpool.Pool
}

func NovoDocumentoRepository(pool *pgxpool.Pool) *DocumentoRepository {
	return &DocumentoRepository{pool: pool}
}

func (r *DocumentoRepository) Salvar(ctx context.Context, d domain.Documento) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO documentos (`+colunasDocumento+`) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		d.ID, string(d.Tipo), d.PacienteID, d.ConsultaID, d.AutorID, d.NomeOriginal,
		d.ContentType, d.TamanhoBytes, d.URI, d.EnviadoEm,
	)
	if err != nil {
		return fmt.Errorf("inserindo documento: %w", err)
	}
	return nil
}

func (r *DocumentoRepository) BuscarPorID(ctx context.Context, id uuid.UUID) (domain.Documento, error) {
	linha := r.pool.QueryRow(ctx, `SELECT `+colunasDocumento+` FROM documentos WHERE id = $1`, id)
	d, err := lerDocumento(linha)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Documento{}, domain.ErrNaoEncontrado
	}
	return d, err
}

func (r *DocumentoRepository) Listar(ctx context.Context, f domain.FiltroDocumentos) ([]domain.Documento, error) {
	var (
		condicoes []string
		args      []any
	)
	adicionar := func(condicao string, valor any) {
		args = append(args, valor)
		condicoes = append(condicoes, fmt.Sprintf(condicao, len(args)))
	}
	if f.PacienteID != nil {
		adicionar("paciente_id = $%d", *f.PacienteID)
	}
	if f.ConsultaID != nil {
		adicionar("consulta_id = $%d", *f.ConsultaID)
	}
	if f.Tipo != nil {
		adicionar("tipo = $%d", string(*f.Tipo))
	}

	sql := `SELECT ` + colunasDocumento + ` FROM documentos`
	if len(condicoes) > 0 {
		sql += ` WHERE ` + strings.Join(condicoes, ` AND `)
	}
	sql += fmt.Sprintf(` ORDER BY enviado_em DESC LIMIT %d`, limiteListagem)

	linhas, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("listando documentos: %w", err)
	}
	docs, err := pgx.CollectRows(linhas, func(l pgx.CollectableRow) (domain.Documento, error) {
		return lerDocumento(l)
	})
	if err != nil {
		return nil, fmt.Errorf("lendo documentos: %w", err)
	}
	return docs, nil
}

func (r *DocumentoRepository) Remover(ctx context.Context, id uuid.UUID) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM documentos WHERE id = $1`, id); err != nil {
		return fmt.Errorf("removendo documento: %w", err)
	}
	return nil
}

type linhaPg interface {
	Scan(dest ...any) error
}

func lerDocumento(l linhaPg) (domain.Documento, error) {
	var (
		d    domain.Documento
		tipo string
	)
	err := l.Scan(&d.ID, &tipo, &d.PacienteID, &d.ConsultaID, &d.AutorID, &d.NomeOriginal,
		&d.ContentType, &d.TamanhoBytes, &d.URI, &d.EnviadoEm)
	d.Tipo = domain.TipoDocumento(tipo)
	d.EnviadoEm = d.EnviadoEm.UTC()
	return d, err
}
