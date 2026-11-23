"""Todas as respostas de erro saem no formato da equipe: {"erro": "mensagem"}."""

import logging

from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from starlette.exceptions import HTTPException as StarletteHTTPException

from app.domain.erros import ErroDeNegocio

MENSAGENS_PADRAO = {
    404: "rota não encontrada",
    405: "método não permitido",
}


def resposta_erro(status: int, mensagem: str, headers: dict[str, str] | None = None) -> JSONResponse:
    return JSONResponse(status_code=status, content={"erro": mensagem}, headers=headers)


def registrar(app: FastAPI, log: logging.Logger) -> None:
    @app.exception_handler(ErroDeNegocio)
    async def erro_de_negocio(_: Request, erro: ErroDeNegocio):
        return resposta_erro(erro.status_http, erro.mensagem)

    @app.exception_handler(StarletteHTTPException)
    async def erro_http(_: Request, erro: StarletteHTTPException):
        mensagem = erro.detail if isinstance(erro.detail, str) else None
        # Troca as mensagens em inglês do Starlette ("Not Found"...) pelas nossas.
        if erro.status_code in MENSAGENS_PADRAO and mensagem in (None, "Not Found", "Method Not Allowed"):
            mensagem = MENSAGENS_PADRAO[erro.status_code]
        return resposta_erro(erro.status_code, mensagem or "erro", getattr(erro, "headers", None))

    @app.exception_handler(RequestValidationError)
    async def erro_validacao(_: Request, erro: RequestValidationError):
        # A equipe usa 400 para dados inválidos (o FastAPI usaria 422).
        return resposta_erro(400, descrever_validacao(erro))


def descrever_validacao(erro: RequestValidationError) -> str:
    """Transforma os erros do Pydantic numa frase curta em português."""
    problemas = []
    for e in erro.errors():
        tipo = e.get("type", "")
        if tipo == "json_invalid":
            return "JSON inválido"
        local = [str(p) for p in e.get("loc", ()) if p not in ("body", "query", "path", "header")]
        campo = ".".join(local)
        if tipo == "missing":
            texto = "obrigatório"
        elif tipo == "uuid_parsing":
            texto = "deve ser um UUID"
        else:
            texto = e.get("msg", "inválido")
        problemas.append(f"{campo}: {texto}" if campo else texto)
    return "; ".join(problemas) or "dados inválidos"
