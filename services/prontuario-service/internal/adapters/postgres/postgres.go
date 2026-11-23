// Package postgres é o adaptador de banco de dados (prontuario_db).
package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var arquivosMigracao embed.FS

// Conectar abre o pool e espera o banco aceitar conexões (no docker-compose o
// Postgres pode levar alguns segundos para subir).
func Conectar(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("configuração do banco inválida: %w", err)
	}

	var errPing error
	for tentativa := 1; tentativa <= 10; tentativa++ {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		errPing = pool.Ping(pingCtx)
		cancel()
		if errPing == nil {
			return pool, nil
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	pool.Close()
	return nil, fmt.Errorf("banco indisponível: %w", errPing)
}

// Migrar aplica, em ordem alfabética, os arquivos de migrations/ que ainda não
// constam na tabela schema_migrations. Cada arquivo roda em uma transação.
func Migrar(ctx context.Context, pool *pgxpool.Pool) error {
	const criarControle = `CREATE TABLE IF NOT EXISTS schema_migrations (
		versao      TEXT PRIMARY KEY,
		aplicada_em TIMESTAMPTZ NOT NULL DEFAULT now()
	)`
	if _, err := pool.Exec(ctx, criarControle); err != nil {
		return fmt.Errorf("criando schema_migrations: %w", err)
	}

	entradas, err := fs.ReadDir(arquivosMigracao, "migrations")
	if err != nil {
		return err
	}
	for _, e := range entradas { // ReadDir já devolve em ordem alfabética
		nome := e.Name()
		if e.IsDir() || !strings.HasSuffix(nome, ".sql") {
			continue
		}
		if err := aplicarMigracao(ctx, pool, nome); err != nil {
			return fmt.Errorf("migração %s: %w", nome, err)
		}
	}
	return nil
}

func aplicarMigracao(ctx context.Context, pool *pgxpool.Pool, nome string) error {
	var aplicada bool
	err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE versao = $1)`, nome).Scan(&aplicada)
	if err != nil || aplicada {
		return err
	}

	sql, err := arquivosMigracao.ReadFile("migrations/" + nome)
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // sem efeito depois do Commit

	// Sem argumentos o pgx usa o protocolo simples, que aceita vários comandos.
	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (versao) VALUES ($1)`, nome); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
