import logging
import time

from fastapi import FastAPI, Request

from .erros import resposta_erro


def registrar_logs(app: FastAPI, log: logging.Logger) -> None:
    """Registra cada requisição (método, caminho, status, duração) e
    transforma erros inesperados em 500 {"erro": "erro interno"}."""

    @app.middleware("http")
    async def logs_e_recuperacao(request: Request, chamar_proximo):
        inicio = time.perf_counter()
        try:
            resposta = await chamar_proximo(request)
        except Exception:
            log.exception("erro inesperado ao tratar %s %s", request.method, request.url.path)
            resposta = resposta_erro(500, "erro interno")
        duracao_ms = round((time.perf_counter() - inicio) * 1000)
        log.info(
            "requisição metodo=%s caminho=%s status=%s duracao_ms=%s",
            request.method, request.url.path, resposta.status_code, duracao_ms,
        )
        return resposta
