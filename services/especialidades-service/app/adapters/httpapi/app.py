"""Monta a aplicação FastAPI: rotas, tratamento de erros, logs e CORS."""

import logging
from dataclasses import dataclass

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.ports import VerificadorBanco

from . import erros, middleware, rotas_health


@dataclass
class Dependencias:
    servico: str
    banco: VerificadorBanco
    log: logging.Logger


def criar_app(dep: Dependencias, cors_origens: list[str]) -> FastAPI:
    app = FastAPI(
        title=dep.servico,
        description="Especialidades médicas, configuração da triagem e cálculo de exames faltantes.",
        version="0.1.0",
    )
    app.state.dep = dep

    app.include_router(rotas_health.router)
    erros.registrar(app, dep.log)

    # Ordem: o último middleware adicionado é o mais externo.
    # Logs (com captura de erro inesperado) ficam por dentro; CORS por fora,
    # para que até as respostas de erro cheguem ao navegador.
    middleware.registrar_logs(app, dep.log)
    app.add_middleware(
        CORSMiddleware,
        allow_origins=cors_origens,
        allow_methods=["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
        allow_headers=["Authorization", "Content-Type"],
        expose_headers=["Content-Disposition"],
    )
    return app
