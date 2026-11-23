import logging
import uuid

import pytest
from fastapi.testclient import TestClient
from pydantic import BaseModel

from app.adapters.httpapi import Dependencias, criar_app
from datetime import datetime, timezone

from app.adapters.memoria import BancoMemoria, EspecialidadeRepositoryMemoria
from app.domain.erros import NaoEncontrado
from app.domain.iniciais import ESPECIALIDADES_INICIAIS
from app.servico.especialidades import ServicoEspecialidades

ORIGEM_FRONT = "http://localhost:5173"


class CorpoTeste(BaseModel):
    nome: str
    medico_id: uuid.UUID


AGORA = datetime(2026, 11, 23, 12, 0, tzinfo=timezone.utc)


def novo_servico() -> ServicoEspecialidades:
    servico = ServicoEspecialidades(EspecialidadeRepositoryMemoria(), relogio=lambda: AGORA)
    servico.semear(ESPECIALIDADES_INICIAIS)
    return servico


def criar_cliente(banco_fora: bool = False) -> TestClient:
    app = criar_app(
        Dependencias(
            servico="especialidades-service",
            banco=BancoMemoria(falhar=banco_fora),
            especialidades=novo_servico(),
            log=logging.getLogger("teste"),
        ),
        [ORIGEM_FRONT],
    )

    # Rotas só de teste, para exercitar o tratamento de erros.
    @app.post("/_teste/corpo")
    def corpo(c: CorpoTeste):
        return c

    @app.get("/_teste/nao-encontrado")
    def nao_encontrado():
        raise NaoEncontrado("especialidade não encontrada")

    @app.get("/_teste/quebra")
    def quebra():
        raise RuntimeError("bug")

    return TestClient(app)


@pytest.fixture
def cliente() -> TestClient:
    return criar_cliente()
