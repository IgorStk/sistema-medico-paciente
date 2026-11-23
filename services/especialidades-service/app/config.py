"""Lê a configuração do serviço a partir de variáveis de ambiente."""

import os
from dataclasses import dataclass, field
from typing import Mapping


@dataclass(frozen=True)
class Config:
    porta: int = 8083
    database_url: str = "postgres://especialidades:especialidades@localhost:5435/especialidades_db?sslmode=disable"
    cors_origens: list[str] = field(default_factory=lambda: ["http://localhost:5173"])

    @property
    def usa_memoria(self) -> bool:
        """DATABASE_URL=memoria roda sem Postgres (dados somem ao reiniciar)."""
        return self.database_url.strip().lower() == "memoria"


def carregar(ambiente: Mapping[str, str] | None = None) -> Config:
    """Lê as variáveis de ambiente. Os valores padrão servem para rodar localmente."""
    ambiente = os.environ if ambiente is None else ambiente
    padrao = Config()

    porta_texto = _env(ambiente, "PORT", str(padrao.porta))
    try:
        porta = int(porta_texto)
    except ValueError:
        raise ValueError(f"PORT inválida: {porta_texto!r}") from None
    if not 1 <= porta <= 65535:
        raise ValueError(f"PORT fora do intervalo 1-65535: {porta}")

    return Config(
        porta=porta,
        database_url=_env(ambiente, "DATABASE_URL", padrao.database_url),
        cors_origens=_separar_lista(_env(ambiente, "CORS_ORIGINS", ",".join(padrao.cors_origens))),
    )


def _env(ambiente: Mapping[str, str], chave: str, padrao: str) -> str:
    valor = ambiente.get(chave, "").strip()
    return valor or padrao


def _separar_lista(texto: str) -> list[str]:
    return [p.strip() for p in texto.split(",") if p.strip()]
