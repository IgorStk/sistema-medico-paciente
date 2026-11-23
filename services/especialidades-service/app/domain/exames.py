"""Cálculo dos exames faltantes de uma consulta."""

from dataclasses import dataclass

from .especialidade import Especialidade
from .texto import normalizar


@dataclass(frozen=True)
class ExameEntregue:
    codigo: str
    nome: str
    documento: str  # nome do arquivo que atendeu o exame


@dataclass(frozen=True)
class ExameFaltante:
    codigo: str
    nome: str
    obrigatorio: bool


@dataclass(frozen=True)
class ResultadoExames:
    entregues: list[ExameEntregue]
    faltantes: list[ExameFaltante]

    @property
    def completo(self) -> bool:
        """True quando nenhum exame obrigatório está faltando."""
        return not any(f.obrigatorio for f in self.faltantes)


def calcular_faltantes(especialidade: Especialidade, documentos: list[str]) -> ResultadoExames:
    """Compara os documentos enviados (nomes de arquivo ou descrições) com os
    documentos esperados pela especialidade.

    Um esperado conta como entregue se algum documento contém, como palavra(s)
    inteira(s), o código, o nome ou uma palavra-chave dele. Ex.: "ECG-joao.pdf"
    atende o eletrocardiograma que tem a palavra-chave "ecg".
    """
    enviados = [(d, f" {normalizar(d)} ") for d in documentos if d.strip()]
    entregues: list[ExameEntregue] = []
    faltantes: list[ExameFaltante] = []

    for esperado in especialidade.documentos_esperados:
        termos = [f" {t} " for t in esperado.termos_de_busca()]
        achado = next((original for original, texto in enviados if any(t in texto for t in termos)), None)
        if achado is None:
            faltantes.append(ExameFaltante(esperado.codigo, esperado.nome, esperado.obrigatorio))
        else:
            entregues.append(ExameEntregue(esperado.codigo, esperado.nome, achado))
    return ResultadoExames(entregues, faltantes)
