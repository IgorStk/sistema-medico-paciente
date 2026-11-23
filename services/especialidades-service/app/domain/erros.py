"""Erros de negócio. A camada HTTP traduz cada um para o status certo."""


class ErroDeNegocio(Exception):
    """Base dos erros que viram resposta {"erro": "mensagem"}."""

    status_http = 400

    def __init__(self, mensagem: str):
        super().__init__(mensagem)
        self.mensagem = mensagem


class DadosInvalidos(ErroDeNegocio):
    status_http = 400


class NaoEncontrado(ErroDeNegocio):
    status_http = 404


class Conflito(ErroDeNegocio):
    status_http = 409
