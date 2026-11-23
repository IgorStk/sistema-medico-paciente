import logging
import time
from pathlib import Path

import psycopg
from psycopg_pool import ConnectionPool

PASTA_MIGRACOES = Path(__file__).parent / "migrations"


def conectar(url: str, log: logging.Logger, tentativas: int = 10) -> ConnectionPool:
    """Espera o banco aceitar conexões (no docker-compose o Postgres pode levar
    alguns segundos para subir) e abre o pool."""
    # prepare_threshold=None: sem prepared statements, para funcionar também
    # com o pooler do Supabase em modo transação (porta 6543).
    opcoes = {"prepare_threshold": None, "connect_timeout": 5}
    for tentativa in range(1, tentativas + 1):
        try:
            psycopg.connect(url, **opcoes).close()
            break
        except psycopg.OperationalError as erro:
            if tentativa == tentativas:
                raise ConnectionError(f"banco indisponível: {erro}") from erro
            log.warning("banco indisponível (tentativa %d de %d): %s", tentativa, tentativas, erro)
            time.sleep(1)

    pool = ConnectionPool(url, min_size=1, max_size=5, open=False, kwargs=opcoes)
    pool.open(wait=True, timeout=10)
    return pool


def migrar(pool: ConnectionPool, log: logging.Logger) -> None:
    """Aplica, em ordem alfabética, os arquivos de migrations/ que ainda não
    constam na tabela schema_migrations. Cada arquivo roda em uma transação."""
    with pool.connection() as conn:
        conn.execute(
            """CREATE TABLE IF NOT EXISTS schema_migrations (
                versao      TEXT PRIMARY KEY,
                aplicada_em TIMESTAMPTZ NOT NULL DEFAULT now()
            )"""
        )
    for arquivo in sorted(PASTA_MIGRACOES.glob("*.sql")):
        with pool.connection() as conn, conn.transaction():
            ja = conn.execute("SELECT 1 FROM schema_migrations WHERE versao = %s", (arquivo.name,)).fetchone()
            if ja:
                continue
            conn.execute(arquivo.read_text(encoding="utf-8"))
            conn.execute("INSERT INTO schema_migrations (versao) VALUES (%s)", (arquivo.name,))
            log.info("migração aplicada: %s", arquivo.name)


class BancoPostgres:
    """Implementa VerificadorBanco para o /health."""

    def __init__(self, pool: ConnectionPool):
        self.pool = pool

    def ping(self) -> None:
        with self.pool.connection(timeout=2) as conn:
            conn.execute("SELECT 1")
