"""Adaptadores em memória: usados nos testes e para rodar sem banco
(DATABASE_URL=memoria)."""

import threading
import uuid

from app.domain.especialidade import Especialidade


class BancoMemoria:
    """Banco falso para o /health. Com `falhar=True`, simula banco fora do ar."""

    def __init__(self, falhar: bool = False):
        self.falhar = falhar

    def ping(self) -> None:
        if self.falhar:
            raise ConnectionError("banco indisponível")


class EspecialidadeRepositoryMemoria:
    def __init__(self):
        self._itens: dict[uuid.UUID, Especialidade] = {}
        self._trava = threading.Lock()

    def listar(self, apenas_ativas: bool) -> list[Especialidade]:
        with self._trava:
            itens = [e for e in self._itens.values() if e.ativa or not apenas_ativas]
        return sorted(itens, key=lambda e: e.nome)

    def buscar_por_id(self, id: uuid.UUID) -> Especialidade | None:
        with self._trava:
            return self._itens.get(id)

    def buscar_por_codigo(self, codigo: str) -> Especialidade | None:
        with self._trava:
            return next((e for e in self._itens.values() if e.codigo == codigo), None)

    def salvar_configuracao(self, especialidade: Especialidade) -> None:
        with self._trava:
            self._itens[especialidade.id] = especialidade

    def inserir_se_nao_existir(self, especialidade: Especialidade) -> bool:
        with self._trava:
            if any(e.codigo == especialidade.codigo or e.id == especialidade.id for e in self._itens.values()):
                return False
            self._itens[especialidade.id] = especialidade
            return True
