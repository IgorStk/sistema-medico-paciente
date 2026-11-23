"""Monta as dependências e cria a aplicação."""

import logging

from fastapi import FastAPI

from app import NOME_SERVICO
from app.adapters.httpapi import Dependencias, criar_app
from app.adapters.memoria import BancoMemoria, EspecialidadeRepositoryMemoria
from app.config import Config
from app.domain.iniciais import ESPECIALIDADES_INICIAIS
from app.servico.especialidades import ServicoEspecialidades


def configurar_logs() -> logging.Logger:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s")
    return logging.getLogger(NOME_SERVICO)


def montar(cfg: Config) -> FastAPI:
    log = configurar_logs()

    if cfg.usa_memoria:
        log.warning("DATABASE_URL=memoria: os dados ficam só na memória e somem ao reiniciar")
        banco, repo = BancoMemoria(), EspecialidadeRepositoryMemoria()
    else:
        # Importado só aqui para quem roda em memória não precisar do psycopg.
        from app.adapters.postgres import BancoPostgres, EspecialidadeRepositoryPostgres, conectar, migrar

        pool = conectar(cfg.database_url, log)
        migrar(pool, log)
        log.info("banco conectado e migrações aplicadas")
        banco, repo = BancoPostgres(pool), EspecialidadeRepositoryPostgres(pool)

    servico = ServicoEspecialidades(repo)
    if inseridas := servico.semear(ESPECIALIDADES_INICIAIS):
        log.info("%d especialidades iniciais cadastradas", inseridas)

    dep = Dependencias(servico=NOME_SERVICO, banco=banco, especialidades=servico, log=log)
    return criar_app(dep, cfg.cors_origens)
