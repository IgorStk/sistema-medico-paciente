from tests.conftest import criar_cliente


def test_health_ok(cliente):
    r = cliente.get("/health")
    assert r.status_code == 200
    assert r.json() == {"status": "ok", "servico": "especialidades-service", "banco": "ok"}


def test_health_banco_fora():
    r = criar_cliente(banco_fora=True).get("/health")
    assert r.status_code == 503
    assert r.json() == {"status": "degradado", "servico": "especialidades-service", "banco": "indisponivel"}
