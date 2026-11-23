"""Formato JSON da API (camelCase, como nos serviços em Go)."""

import uuid
from datetime import datetime

from pydantic import BaseModel, ConfigDict, Field
from pydantic.alias_generators import to_camel

from app.domain.especialidade import DocumentoEsperado, Especialidade
from app.domain.exames import ResultadoExames


class Modelo(BaseModel):
    model_config = ConfigDict(alias_generator=to_camel, populate_by_name=True)


class DocumentoEsperadoJSON(Modelo):
    codigo: str = Field(examples=["eletrocardiograma"])
    nome: str = Field(examples=["Eletrocardiograma (ECG)"])
    obrigatorio: bool = True
    palavras_chave: list[str] = Field(default_factory=list, examples=[["ecg", "eletro"]])

    @classmethod
    def de(cls, d: DocumentoEsperado) -> "DocumentoEsperadoJSON":
        return cls(codigo=d.codigo, nome=d.nome, obrigatorio=d.obrigatorio, palavras_chave=list(d.palavras_chave))

    def para_dominio(self) -> DocumentoEsperado:
        return DocumentoEsperado(self.codigo, self.nome, self.obrigatorio, tuple(self.palavras_chave))


class EspecialidadeResumo(Modelo):
    id: uuid.UUID
    codigo: str
    nome: str
    descricao: str
    ativa: bool

    @classmethod
    def de(cls, e: Especialidade) -> "EspecialidadeResumo":
        return cls(id=e.id, codigo=e.codigo, nome=e.nome, descricao=e.descricao, ativa=e.ativa)


class Configuracao(Modelo):
    especialidade_id: uuid.UUID
    prompt_base: str
    documentos_esperados: list[DocumentoEsperadoJSON]
    atualizado_em: datetime | None

    @classmethod
    def de(cls, e: Especialidade) -> "Configuracao":
        return cls(
            especialidade_id=e.id,
            prompt_base=e.prompt_base,
            documentos_esperados=[DocumentoEsperadoJSON.de(d) for d in e.documentos_esperados],
            atualizado_em=e.atualizado_em,
        )


class EspecialidadeDetalhe(EspecialidadeResumo):
    prompt_base: str
    documentos_esperados: list[DocumentoEsperadoJSON]
    atualizado_em: datetime | None

    @classmethod
    def de(cls, e: Especialidade) -> "EspecialidadeDetalhe":
        return cls(
            **EspecialidadeResumo.de(e).model_dump(),
            prompt_base=e.prompt_base,
            documentos_esperados=[DocumentoEsperadoJSON.de(d) for d in e.documentos_esperados],
            atualizado_em=e.atualizado_em,
        )


class AtualizarConfiguracao(Modelo):
    prompt_base: str = Field(examples=["Você apoia o médico na triagem..."])
    documentos_esperados: list[DocumentoEsperadoJSON]


class PedidoExamesFaltantes(Modelo):
    documentos: list[str] = Field(
        description="Nomes dos arquivos (ou descrições) já enviados para a consulta.",
        examples=[["hemograma-joao.pdf", "ECG_2026.pdf"]],
    )


class ExameEntregueJSON(Modelo):
    codigo: str
    nome: str
    documento: str


class ExameFaltanteJSON(Modelo):
    codigo: str
    nome: str
    obrigatorio: bool


class RespostaExamesFaltantes(Modelo):
    especialidade_id: uuid.UUID
    completo: bool = Field(description="true quando não falta nenhum exame obrigatório")
    entregues: list[ExameEntregueJSON]
    faltantes: list[ExameFaltanteJSON]

    @classmethod
    def de(cls, e: Especialidade, r: ResultadoExames) -> "RespostaExamesFaltantes":
        return cls(
            especialidade_id=e.id,
            completo=r.completo,
            entregues=[ExameEntregueJSON(codigo=x.codigo, nome=x.nome, documento=x.documento) for x in r.entregues],
            faltantes=[ExameFaltanteJSON(codigo=x.codigo, nome=x.nome, obrigatorio=x.obrigatorio) for x in r.faltantes],
        )
