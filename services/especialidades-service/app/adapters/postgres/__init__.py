"""Adaptador de banco de dados (especialidades_db) com psycopg 3."""

from .postgres import BancoPostgres, conectar, migrar
from .especialidade_repository import EspecialidadeRepositoryPostgres

__all__ = ["BancoPostgres", "EspecialidadeRepositoryPostgres", "conectar", "migrar"]
