"""Testes do repositório Postgres. Só rodam com TEST_DATABASE_URL apontando
para um banco DESCARTÁVEL (as tabelas são apagadas antes do teste)."""

import logging
import os
from dataclasses import replace
from datetime import datetime, timezone

import pytest

URL = os.environ.get("TEST_DATABASE_URL")
pytestmark = pytest.mark.skipif(not URL, reason="defina TEST_DATABASE_URL para testar com Postgres")

LOG = logging.getLogger("teste")


@pytest.fixture
def pool():
    from app.adapters.postgres import conectar

    pool = conectar(URL, LOG, tentativas=1)
    with pool.connection() as conn:
        conn.execute("DROP TABLE IF EXISTS documentos_esperados, especialidades, schema_migrations")
    yield pool
    pool.close()


def test_repositorio_postgres(pool):
    from app.adapters.postgres import BancoPostgres, EspecialidadeRepositoryPostgres, migrar
    from app.domain.especialidade import DocumentoEsperado
    from app.domain.iniciais import ESPECIALIDADES_INICIAIS
    from app.servico.especialidades import ServicoEspecialidades

    migrar(pool, LOG)
    migrar(pool, LOG)  # rodar de novo não faz nada
    BancoPostgres(pool).ping()

    agora = datetime(2026, 11, 23, 12, 0, tzinfo=timezone.utc)
    servico = ServicoEspecialidades(EspecialidadeRepositoryPostgres(pool), relogio=lambda: agora)
    assert servico.semear(ESPECIALIDADES_INICIAIS) == len(ESPECIALIDADES_INICIAIS)
    assert servico.semear(ESPECIALIDADES_INICIAIS) == 0

    cardio = servico.obter("cardiologia")
    esperado = next(e for e in ESPECIALIDADES_INICIAIS if e.codigo == "cardiologia")
    assert cardio == replace(esperado, atualizado_em=agora)
    assert servico.obter(str(cardio.id)) == cardio
    assert [e.nome for e in servico.listar()] == sorted(e.nome for e in ESPECIALIDADES_INICIAIS)

    servico.atualizar_configuracao("cardiologia", "novo", [DocumentoEsperado("holter", "Holter", False, ("holter",))])
    salva = servico.obter("cardiologia")
    assert salva.prompt_base == "novo"
    assert salva.documentos_esperados == (DocumentoEsperado("holter", "Holter", False, ("holter",)),)

    # Semear de novo não sobrescreve o que foi editado.
    servico.semear(ESPECIALIDADES_INICIAIS)
    assert servico.obter("cardiologia").prompt_base == "novo"
