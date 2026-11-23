"""Portas (interfaces) que os adaptadores implementam."""

from typing import Protocol


class VerificadorBanco(Protocol):
    """Usado pelo /health. Lança exceção se o banco não responder."""

    def ping(self) -> None: ...
