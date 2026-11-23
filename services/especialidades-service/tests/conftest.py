import logging
import uuid

import pytest
from fastapi.testclient import TestClient
from pydantic import BaseModel

from app.adapters.httpapi import Dependencias, criar_app
from app.adapters.memoria import BancoMemoria
from app.domain.erros import NaoEncontrado

ORIGEM_FRONT = "http://localhost:5173"


class CorpoTeste(BaseModel):
    nome: str
    medico_id: uuid.UUID


def criar_cliente(banco_fora: bool = False) -> TestClient:
    app = criar_app(
        Dependencias(servico="especialidades-service", banco=BancoMemoria(falhar=banco_fora), log=logging.getLogger("teste")),
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
