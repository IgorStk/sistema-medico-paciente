"""Monta as dependências e cria a aplicação."""

import logging

from fastapi import FastAPI

from app import NOME_SERVICO
from app.adapters.httpapi import Dependencias, criar_app
from app.adapters.memoria import BancoMemoria
from app.config import Config


def configurar_logs() -> logging.Logger:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s")
    return logging.getLogger(NOME_SERVICO)


def montar(cfg: Config) -> FastAPI:
    log = configurar_logs()
    # Etapa 1: ainda sem Postgres. Na Etapa 2 o BancoMemoria é trocado pelo
    # adaptador do Postgres (DATABASE_URL).
    banco = BancoMemoria()
    log.info("usando armazenamento em memória (Postgres entra na Etapa 2)")
    return criar_app(Dependencias(servico=NOME_SERVICO, banco=banco, log=log), cfg.cors_origens)
