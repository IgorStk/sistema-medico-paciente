import pytest

from app.config import carregar


def test_padroes():
    cfg = carregar({})
    assert cfg.porta == 8083
    assert cfg.database_url.startswith("postgres://especialidades:")
    assert cfg.cors_origens == ["http://localhost:5173"]


def test_le_variaveis():
    cfg = carregar({
        "PORT": " 9000 ",
        "DATABASE_URL": "postgresql://u:s@host:5432/db",
        "CORS_ORIGINS": "http://localhost:5173, http://localhost:3000,,",
    })
    assert cfg.porta == 9000
    assert cfg.database_url == "postgresql://u:s@host:5432/db"
    assert cfg.cors_origens == ["http://localhost:5173", "http://localhost:3000"]


def test_vazio_usa_padrao():
    assert carregar({"PORT": "  ", "CORS_ORIGINS": ""}).porta == 8083


@pytest.mark.parametrize("porta", ["abc", "0", "70000"])
def test_porta_invalida(porta):
    with pytest.raises(ValueError):
        carregar({"PORT": porta})
