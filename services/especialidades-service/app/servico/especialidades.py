"""Casos de uso do serviço de especialidades."""

import uuid
from dataclasses import replace
from datetime import datetime, timezone
from typing import Callable

from app.domain.erros import NaoEncontrado
from app.domain.especialidade import DocumentoEsperado, Especialidade
from app.domain.exames import ResultadoExames, calcular_faltantes
from app.ports import EspecialidadeRepository


class ServicoEspecialidades:
    def __init__(self, repo: EspecialidadeRepository, relogio: Callable[[], datetime] | None = None):
        self.repo = repo
        self.relogio = relogio or (lambda: datetime.now(timezone.utc))

    def listar(self, apenas_ativas: bool = True) -> list[Especialidade]:
        return self.repo.listar(apenas_ativas)

    def obter(self, ref: str) -> Especialidade:
        """`ref` pode ser o UUID ou o código (ex.: "cardiologia")."""
        try:
            encontrada = self.repo.buscar_por_id(uuid.UUID(ref))
        except ValueError:
            encontrada = self.repo.buscar_por_codigo(ref.strip().lower())
        if encontrada is None:
            raise NaoEncontrado("especialidade não encontrada")
        return encontrada

    def atualizar_configuracao(self, ref: str, prompt_base: str, documentos: list[DocumentoEsperado]) -> Especialidade:
        atualizada = self.obter(ref).com_configuracao(prompt_base, documentos, self.relogio())
        self.repo.salvar_configuracao(atualizada)
        return atualizada

    def exames_faltantes(self, ref: str, documentos: list[str]) -> tuple[Especialidade, ResultadoExames]:
        especialidade = self.obter(ref)
        return especialidade, calcular_faltantes(especialidade, documentos)

    def semear(self, iniciais: list[Especialidade]) -> int:
        """Cadastra as especialidades iniciais que ainda não existem (não
        sobrescreve o que já foi editado). Devolve quantas inseriu."""
        agora = self.relogio()
        return sum(
            self.repo.inserir_se_nao_existir(e if e.atualizado_em else _com_data(e, agora)) for e in iniciais
        )


def _com_data(e: Especialidade, agora: datetime) -> Especialidade:
    return replace(e, atualizado_em=agora)
