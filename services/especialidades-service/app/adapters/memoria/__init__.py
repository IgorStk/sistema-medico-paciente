"""Adaptadores em memória: usados nos testes e para rodar sem banco."""


class BancoMemoria:
    """Banco falso para o /health. Com `falhar=True`, simula banco fora do ar."""

    def __init__(self, falhar: bool = False):
        self.falhar = falhar

    def ping(self) -> None:
        if self.falhar:
            raise ConnectionError("banco indisponível")
