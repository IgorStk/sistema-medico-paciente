"""Especialidades cadastradas automaticamente na primeira subida.

Os IDs são fixos para que os outros serviços (agendamento, triagem, frontend)
possam usá-los nos testes. Depois de cadastradas, a configuração pode ser
editada pela API (PUT /especialidades/{id}/configuracao) sem ser sobrescrita.
"""

import uuid

from .especialidade import DocumentoEsperado as Doc
from .especialidade import Especialidade

_REGRAS = (
    "Você apoia o médico na triagem antes da consulta. Responda em português do Brasil, "
    "de forma objetiva. Não dê diagnóstico definitivo nem prescreva: escreva um resumo "
    "clínico, um rascunho de conduta para o médico revisar e aponte exames que faltam. "
)

ESPECIALIDADES_INICIAIS: list[Especialidade] = [
    Especialidade(
        id=uuid.UUID("3a1f0c1e-0001-4000-8000-000000000001"),
        codigo="clinica-geral",
        nome="Clínica Geral",
        descricao="Atendimento inicial, check-up e acompanhamento de doenças comuns.",
        prompt_base=_REGRAS + "Foque em queixa principal, sinais de alarme e exames de rotina.",
        documentos_esperados=(
            Doc("hemograma", "Hemograma completo", True, ("hemograma",)),
            Doc("glicemia", "Glicemia de jejum", True, ("glicemia", "glicose")),
            Doc("perfil-lipidico", "Perfil lipídico", False, ("lipidograma", "colesterol", "lipidico")),
            Doc("urina", "Exame de urina (EAS)", False, ("eas", "urina", "sumario de urina")),
        ),
    ),
    Especialidade(
        id=uuid.UUID("3a1f0c1e-0002-4000-8000-000000000002"),
        codigo="cardiologia",
        nome="Cardiologia",
        descricao="Coração e sistema circulatório.",
        prompt_base=_REGRAS + "Foque em dor torácica, falta de ar, palpitações, pressão arterial e fatores de risco cardiovascular.",
        documentos_esperados=(
            Doc("eletrocardiograma", "Eletrocardiograma (ECG)", True, ("ecg", "eletro")),
            Doc("ecocardiograma", "Ecocardiograma", False, ("eco", "ecocardio")),
            Doc("hemograma", "Hemograma completo", True, ("hemograma",)),
            Doc("perfil-lipidico", "Perfil lipídico", True, ("lipidograma", "colesterol", "lipidico")),
            Doc("teste-ergometrico", "Teste ergométrico", False, ("ergometrico", "ergometria", "teste de esforco")),
        ),
    ),
    Especialidade(
        id=uuid.UUID("3a1f0c1e-0003-4000-8000-000000000003"),
        codigo="endocrinologia",
        nome="Endocrinologia",
        descricao="Hormônios, diabetes, tireoide e metabolismo.",
        prompt_base=_REGRAS + "Foque em controle glicêmico, tireoide, peso e sintomas hormonais.",
        documentos_esperados=(
            Doc("glicemia", "Glicemia de jejum", True, ("glicemia", "glicose")),
            Doc("hemoglobina-glicada", "Hemoglobina glicada (HbA1c)", True, ("hba1c", "glicada")),
            Doc("tsh", "TSH", True, ("tsh",)),
            Doc("t4-livre", "T4 livre", False, ("t4",)),
            Doc("perfil-lipidico", "Perfil lipídico", False, ("lipidograma", "colesterol", "lipidico")),
        ),
    ),
    Especialidade(
        id=uuid.UUID("3a1f0c1e-0004-4000-8000-000000000004"),
        codigo="ortopedia",
        nome="Ortopedia",
        descricao="Ossos, articulações, músculos e coluna.",
        prompt_base=_REGRAS + "Foque em local e tempo da dor, trauma, limitação de movimento e exames de imagem.",
        documentos_esperados=(
            Doc("raio-x", "Raio-X da região", True, ("rx", "raio x", "radiografia")),
            Doc("ressonancia", "Ressonância magnética", False, ("rm", "ressonancia magnetica")),
        ),
    ),
    Especialidade(
        id=uuid.UUID("3a1f0c1e-0005-4000-8000-000000000005"),
        codigo="dermatologia",
        nome="Dermatologia",
        descricao="Pele, cabelos e unhas.",
        prompt_base=_REGRAS + "Foque em aspecto, local e evolução da lesão, coceira, dor e exposição ao sol.",
        documentos_esperados=(
            Doc("foto-lesao", "Foto da lesão", True, ("foto", "imagem", "lesao")),
            Doc("biopsia", "Resultado de biópsia", False, ("biopsia", "histopatologico")),
        ),
    ),
]
