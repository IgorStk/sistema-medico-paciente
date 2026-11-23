from tests.conftest import criar_cliente

CARDIO_ID = "3a1f0c1e-0002-4000-8000-000000000002"


def test_listar_em_ordem_alfabetica_e_camel_case(cliente):
    r = cliente.get("/especialidades")
    assert r.status_code == 200
    nomes = [e["nome"] for e in r.json()]
    assert nomes == sorted(nomes) and "Cardiologia" in nomes
    assert set(r.json()[0]) == {"id", "codigo", "nome", "descricao", "ativa"}


def test_obter_por_codigo_ou_uuid(cliente):
    por_codigo = cliente.get("/especialidades/Cardiologia")
    por_id = cliente.get(f"/especialidades/{CARDIO_ID}")
    assert por_codigo.status_code == por_id.status_code == 200
    assert por_codigo.json() == por_id.json()
    corpo = por_id.json()
    assert corpo["id"] == CARDIO_ID
    assert corpo["promptBase"]
    assert corpo["documentosEsperados"][0] == {
        "codigo": "eletrocardiograma",
        "nome": "Eletrocardiograma (ECG)",
        "obrigatorio": True,
        "palavrasChave": ["ecg", "eletro"],
    }


def test_inexistente_da_404(cliente):
    for ref in ("pediatria", "00000000-0000-4000-8000-000000000000"):
        r = cliente.get(f"/especialidades/{ref}/configuracao")
        assert r.status_code == 404
        assert r.json() == {"erro": "especialidade não encontrada"}


def test_atualizar_configuracao():
    cliente = criar_cliente()
    nova = {
        "promptBase": "Novo prompt da cardiologia",
        "documentosEsperados": [
            {"codigo": "holter", "nome": "Holter 24h", "palavrasChave": ["holter"]},
            {"codigo": "ecg", "nome": "ECG", "obrigatorio": False},
        ],
    }
    r = cliente.put("/especialidades/cardiologia/configuracao", json=nova)
    assert r.status_code == 200, r.text
    corpo = r.json()
    assert corpo["especialidadeId"] == CARDIO_ID
    assert corpo["promptBase"] == "Novo prompt da cardiologia"
    assert [d["codigo"] for d in corpo["documentosEsperados"]] == ["holter", "ecg"]
    assert corpo["documentosEsperados"][0]["obrigatorio"] is True  # padrão
    assert corpo["atualizadoEm"] == "2026-11-23T12:00:00Z"

    # A leitura seguinte já vem com a configuração nova.
    assert cliente.get(f"/especialidades/{CARDIO_ID}/configuracao").json() == corpo


def test_atualizar_configuracao_invalida():
    cliente = criar_cliente()
    casos = [
        ({"documentosEsperados": []}, "promptBase: obrigatório"),
        ({"promptBase": "  ", "documentosEsperados": []}, "promptBase: obrigatório"),
        ({"promptBase": "p", "documentosEsperados": [{"codigo": "a b", "nome": "A"}]}, "documentosEsperados.0.codigo"),
        ({"promptBase": "p", "documentosEsperados": [{"codigo": "a", "nome": "A"}, {"codigo": "a", "nome": "B"}]},
         "documentosEsperados: código repetido 'a'"),
        ({"promptBase": "p", "documentosEsperados": [{"codigo": "a"}]}, "documentosEsperados.0.nome: obrigatório"),
    ]
    for corpo, mensagem in casos:
        r = cliente.put("/especialidades/cardiologia/configuracao", json=corpo)
        assert r.status_code == 400, corpo
        assert r.json()["erro"].startswith(mensagem), r.json()

    # Nada foi alterado.
    assert cliente.get("/especialidades/cardiologia").json()["documentosEsperados"][0]["codigo"] == "eletrocardiograma"


def test_exames_faltantes(cliente):
    r = cliente.post(
        "/especialidades/cardiologia/exames-faltantes",
        json={"documentos": ["ecg_maria.pdf", "hemograma.pdf"]},
    )
    assert r.status_code == 200
    corpo = r.json()
    assert corpo["especialidadeId"] == CARDIO_ID
    assert corpo["completo"] is False  # falta o perfil lipídico, que é obrigatório
    assert corpo["entregues"] == [
        {"codigo": "eletrocardiograma", "nome": "Eletrocardiograma (ECG)", "documento": "ecg_maria.pdf"},
        {"codigo": "hemograma", "nome": "Hemograma completo", "documento": "hemograma.pdf"},
    ]
    assert {"codigo": "perfil-lipidico", "nome": "Perfil lipídico", "obrigatorio": True} in corpo["faltantes"]


def test_exames_faltantes_corpo_invalido(cliente):
    r = cliente.post("/especialidades/cardiologia/exames-faltantes", json={})
    assert r.status_code == 400
    assert r.json() == {"erro": "documentos: obrigatório"}
