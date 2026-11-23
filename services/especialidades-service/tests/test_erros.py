import uuid

import pytest

from tests.conftest import ORIGEM_FRONT


@pytest.mark.parametrize(
    "metodo, caminho, corpo, status, mensagem",
    [
        ("GET", "/nao-existe", None, 404, "rota não encontrada"),
        ("DELETE", "/health", None, 405, "método não permitido"),
        ("GET", "/_teste/nao-encontrado", None, 404, "especialidade não encontrada"),
        ("GET", "/_teste/quebra", None, 500, "erro interno"),
        ("POST", "/_teste/corpo", {"medico_id": str(uuid.uuid4())}, 400, "nome: obrigatório"),
        ("POST", "/_teste/corpo", {"nome": "x", "medico_id": "abc"}, 400, "medico_id: deve ser um UUID"),
    ],
)
def test_erros_no_formato_da_equipe(cliente, metodo, caminho, corpo, status, mensagem):
    r = cliente.request(metodo, caminho, json=corpo)
    assert r.status_code == status
    assert r.json() == {"erro": mensagem}


def test_json_quebrado(cliente):
    r = cliente.post("/_teste/corpo", content="{", headers={"Content-Type": "application/json"})
    assert r.status_code == 400
    assert r.json() == {"erro": "JSON inválido"}


def test_cors_libera_o_frontend(cliente):
    r = cliente.options(
        "/health",
        headers={"Origin": ORIGEM_FRONT, "Access-Control-Request-Method": "GET"},
    )
    assert r.status_code == 200
    assert r.headers["access-control-allow-origin"] == ORIGEM_FRONT

    r = cliente.get("/health", headers={"Origin": ORIGEM_FRONT})
    assert r.headers["access-control-allow-origin"] == ORIGEM_FRONT


def test_cors_bloqueia_outras_origens(cliente):
    r = cliente.get("/health", headers={"Origin": "http://site-estranho.com"})
    assert "access-control-allow-origin" not in r.headers


def test_erro_interno_tambem_tem_cors(cliente):
    r = cliente.get("/_teste/quebra", headers={"Origin": ORIGEM_FRONT})
    assert r.status_code == 500
    assert r.headers["access-control-allow-origin"] == ORIGEM_FRONT
