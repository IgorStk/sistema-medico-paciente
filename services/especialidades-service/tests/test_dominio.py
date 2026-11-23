import uuid

import pytest

from app.domain.erros import DadosInvalidos
from app.domain.especialidade import DocumentoEsperado as Doc
from app.domain.especialidade import Especialidade, validar_configuracao
from app.domain.exames import calcular_faltantes
from app.domain.iniciais import ESPECIALIDADES_INICIAIS
from app.domain.texto import normalizar

CARDIO = next(e for e in ESPECIALIDADES_INICIAIS if e.codigo == "cardiologia")


def test_normalizar():
    assert normalizar("Raio-X_Tórax (2).PDF") == "raio x torax 2 pdf"
    assert normalizar("  ") == ""


def test_faltantes_casa_por_palavra_chave_codigo_e_nome():
    r = calcular_faltantes(CARDIO, ["ECG-joao.pdf", "Hemograma completo.pdf", "colesterol_total.jpg"])
    assert {e.codigo: e.documento for e in r.entregues} == {
        "eletrocardiograma": "ECG-joao.pdf",
        "hemograma": "Hemograma completo.pdf",
        "perfil-lipidico": "colesterol_total.jpg",
    }
    assert [f.codigo for f in r.faltantes] == ["ecocardiograma", "teste-ergometrico"]
    assert r.completo  # só faltam opcionais


def test_faltantes_sem_documentos():
    r = calcular_faltantes(CARDIO, [])
    assert r.entregues == []
    assert len(r.faltantes) == len(CARDIO.documentos_esperados)
    assert not r.completo


def test_palavra_chave_precisa_ser_palavra_inteira():
    # "eco" não pode casar com "ecografia-abdominal" nem "eletro" com "eletrolitos".
    r = calcular_faltantes(CARDIO, ["ecografia-abdominal.pdf", "eletrolitos.pdf"])
    assert r.entregues == []


def test_acentos_e_maiusculas_nao_importam():
    endo = next(e for e in ESPECIALIDADES_INICIAIS if e.codigo == "endocrinologia")
    r = calcular_faltantes(endo, ["GLICEMIA DE JEJUM.pdf", "HbA1c.png", "TSH e T4.pdf"])
    assert [e.codigo for e in r.entregues] == ["glicemia", "hemoglobina-glicada", "tsh", "t4-livre"]
    assert r.completo


def test_validar_configuracao_limpa_os_dados():
    prompt, docs = validar_configuracao("  Prompt  ", [Doc(" ECG-12 ", " Eletro ", True, (" ECG ", "ecg", " "))])
    assert prompt == "Prompt"
    assert docs == (Doc("ecg-12", "Eletro", True, ("ecg",)),)


@pytest.mark.parametrize(
    "prompt, docs, mensagem",
    [
        ("", [], "promptBase: obrigatório"),
        ("x" * 8001, [], "promptBase: no máximo"),
        ("p", [Doc("com espaco", "N")], "documentosEsperados.0.codigo"),
        ("p", [Doc("a", "A"), Doc("a", "B")], "código repetido"),
        ("p", [Doc("a", " ")], "documentosEsperados.0.nome: obrigatório"),
        ("p", [Doc(f"d{i}", "N") for i in range(51)], "no máximo 50"),
    ],
)
def test_validar_configuracao_erros(prompt, docs, mensagem):
    with pytest.raises(DadosInvalidos, match=mensagem):
        validar_configuracao(prompt, docs)


def test_iniciais_sao_validas_e_unicas():
    assert len({e.id for e in ESPECIALIDADES_INICIAIS}) == len(ESPECIALIDADES_INICIAIS)
    assert len({e.codigo for e in ESPECIALIDADES_INICIAIS}) == len(ESPECIALIDADES_INICIAIS)
    for e in ESPECIALIDADES_INICIAIS:
        assert validar_configuracao(e.prompt_base, list(e.documentos_esperados))[1] == e.documentos_esperados


def test_com_configuracao_nao_altera_o_original():
    e = Especialidade(uuid.uuid4(), "x", "X", "", "antigo")
    nova = e.com_configuracao("novo", [], agora=None)
    assert e.prompt_base == "antigo" and nova.prompt_base == "novo"
