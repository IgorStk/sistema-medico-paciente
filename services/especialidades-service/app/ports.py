"""Portas (interfaces) que os adaptadores implementam."""

import uuid
from typing import Protocol

from app.domain.especialidade import Especialidade


class VerificadorBanco(Protocol):
    """Usado pelo /health. Lança exceção se o banco não responder."""

    def ping(self) -> None: ...


class EspecialidadeRepository(Protocol):
    def listar(self, apenas_ativas: bool) -> list[Especialidade]: ...

    def buscar_por_id(self, id: uuid.UUID) -> Especialidade | None: ...

    def buscar_por_codigo(self, codigo: str) -> Especialidade | None: ...

    def salvar_configuracao(self, especialidade: Especialidade) -> None:
        """Grava prompt_base, documentos_esperados e atualizado_em."""
        ...

    def inserir_se_nao_existir(self, especialidade: Especialidade) -> bool:
        """Insere se ainda não existe uma especialidade com o mesmo código.
        Devolve True se inseriu."""
        ...
