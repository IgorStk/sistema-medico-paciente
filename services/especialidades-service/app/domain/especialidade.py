"""Especialidade médica e a configuração usada pela triagem."""

import re
import uuid
from dataclasses import dataclass, field, replace
from datetime import datetime

from .erros import DadosInvalidos
from .texto import normalizar

MAX_PROMPT = 8000
MAX_DOCUMENTOS = 50
_CODIGO = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")


@dataclass(frozen=True)
class DocumentoEsperado:
    """Exame ou documento que o médico espera receber antes da consulta.

    `palavras_chave` são termos procurados no nome do arquivo enviado
    (ex.: "ecg" para o eletrocardiograma). O código e o nome também contam.
    """

    codigo: str
    nome: str
    obrigatorio: bool = True
    palavras_chave: tuple[str, ...] = ()

    def termos_de_busca(self) -> list[str]:
        termos = [self.codigo, self.nome, *self.palavras_chave]
        return [t for t in dict.fromkeys(normalizar(t) for t in termos) if t]


@dataclass(frozen=True)
class Especialidade:
    id: uuid.UUID
    codigo: str
    nome: str
    descricao: str
    prompt_base: str
    documentos_esperados: tuple[DocumentoEsperado, ...] = ()
    ativa: bool = True
    atualizado_em: datetime | None = field(default=None, compare=False)

    def com_configuracao(self, prompt_base: str, documentos: list[DocumentoEsperado], agora: datetime) -> "Especialidade":
        prompt, docs = validar_configuracao(prompt_base, documentos)
        return replace(self, prompt_base=prompt, documentos_esperados=docs, atualizado_em=agora)


def validar_codigo(codigo: str, campo: str = "codigo") -> str:
    codigo = codigo.strip().lower()
    if not _CODIGO.match(codigo):
        raise DadosInvalidos(f"{campo}: use letras minúsculas, números e hífen (ex.: raio-x-torax)")
    return codigo


def validar_configuracao(
    prompt_base: str, documentos: list[DocumentoEsperado]
) -> tuple[str, tuple[DocumentoEsperado, ...]]:
    prompt = prompt_base.strip()
    if not prompt:
        raise DadosInvalidos("promptBase: obrigatório")
    if len(prompt) > MAX_PROMPT:
        raise DadosInvalidos(f"promptBase: no máximo {MAX_PROMPT} caracteres")
    if len(documentos) > MAX_DOCUMENTOS:
        raise DadosInvalidos(f"documentosEsperados: no máximo {MAX_DOCUMENTOS} itens")

    limpos: list[DocumentoEsperado] = []
    vistos: set[str] = set()
    for i, d in enumerate(documentos):
        codigo = validar_codigo(d.codigo, f"documentosEsperados.{i}.codigo")
        if codigo in vistos:
            raise DadosInvalidos(f"documentosEsperados: código repetido {codigo!r}")
        vistos.add(codigo)
        nome = d.nome.strip()
        if not nome:
            raise DadosInvalidos(f"documentosEsperados.{i}.nome: obrigatório")
        chaves = tuple(dict.fromkeys(p.strip().lower() for p in d.palavras_chave if p.strip()))
        limpos.append(DocumentoEsperado(codigo, nome, d.obrigatorio, chaves))
    return prompt, tuple(limpos)
