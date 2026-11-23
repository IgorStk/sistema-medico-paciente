# especialidades-service (Python / FastAPI)

Microsserviço das **especialidades médicas**: lista as especialidades, guarda a
configuração de cada uma (prompt base da triagem e documentos esperados) e
calcula os exames que ainda faltam numa consulta. Tem banco próprio:
`especialidades_db` (Postgres, no Supabase em produção).

Quem usa a configuração é o `triagem-service`: ele pega o prompt base da
especialidade, os documentos já enviados (no documentos-service) e pergunta
aqui quais exames faltam.

## Endpoints

Em toda rota, `{ref}` aceita o **UUID** ou o **código** da especialidade
(ex.: `cardiologia`). A documentação interativa fica em `/docs`.

| Método | Rota | Quem chama | Descrição |
|---|---|---|---|
| GET | `/health` | — | Saúde do serviço e do banco (`200` ok, `503` banco fora) |
| GET | `/especialidades` | frontend, agendamento | Lista as ativas em ordem alfabética (`?todas=true` inclui inativas) |
| GET | `/especialidades/{ref}` | frontend | Dados completos, com a configuração |
| GET | `/especialidades/{ref}/configuracao` | triagem-service | Prompt base e documentos esperados |
| PUT | `/especialidades/{ref}/configuracao` | frontend (admin) | Substitui o prompt base e a lista de documentos esperados |
| POST | `/especialidades/{ref}/exames-faltantes` | triagem-service, frontend | Compara os documentos enviados com os esperados |

### Especialidades já cadastradas

Na primeira subida o serviço cadastra 5 especialidades com IDs fixos (úteis
para os outros serviços testarem). Depois disso, editar a configuração pela
API não é desfeito ao reiniciar.

| Código | ID |
|---|---|
| `clinica-geral` | `3a1f0c1e-0001-4000-8000-000000000001` |
| `cardiologia` | `3a1f0c1e-0002-4000-8000-000000000002` |
| `endocrinologia` | `3a1f0c1e-0003-4000-8000-000000000003` |
| `ortopedia` | `3a1f0c1e-0004-4000-8000-000000000004` |
| `dermatologia` | `3a1f0c1e-0005-4000-8000-000000000005` |

### Configuração

```json
PUT /especialidades/cardiologia/configuracao
{
  "promptBase": "Você apoia o médico na triagem antes da consulta...",
  "documentosEsperados": [
    { "codigo": "eletrocardiograma", "nome": "Eletrocardiograma (ECG)", "obrigatorio": true, "palavrasChave": ["ecg", "eletro"] },
    { "codigo": "ecocardiograma", "nome": "Ecocardiograma", "obrigatorio": false, "palavrasChave": ["eco"] }
  ]
}
```

- `codigo`: letras minúsculas, números e hífen; não pode repetir.
- `obrigatorio`: padrão `true`.
- `palavrasChave`: termos procurados no nome do arquivo enviado.

A resposta (igual à do `GET .../configuracao`) traz também `especialidadeId`
e `atualizadoEm`.

### Exames faltantes

```json
POST /especialidades/cardiologia/exames-faltantes
{ "documentos": ["ECG_maria.pdf", "hemograma.pdf"] }
```

```json
{
  "especialidadeId": "3a1f0c1e-0002-4000-8000-000000000002",
  "completo": false,
  "entregues": [
    { "codigo": "eletrocardiograma", "nome": "Eletrocardiograma (ECG)", "documento": "ECG_maria.pdf" },
    { "codigo": "hemograma", "nome": "Hemograma completo", "documento": "hemograma.pdf" }
  ],
  "faltantes": [
    { "codigo": "ecocardiograma", "nome": "Ecocardiograma", "obrigatorio": false },
    { "codigo": "perfil-lipidico", "nome": "Perfil lipídico", "obrigatorio": true },
    { "codigo": "teste-ergometrico", "nome": "Teste ergométrico", "obrigatorio": false }
  ]
}
```

Um documento atende um exame esperado quando o nome dele contém, como palavra
inteira, o código, o nome ou uma palavra-chave do exame. Maiúsculas, acentos e
separadores (`-`, `_`, `.`) não importam: `ECG_maria.pdf` atende "ecg", mas
`ecografia.pdf` não atende "eco". `completo` é `true` quando não falta nenhum
exame **obrigatório**.

Os erros vêm no formato da equipe, `{"erro": "mensagem"}`: `400` para dados
inválidos (inclusive JSON quebrado), `404` para especialidade ou rota
inexistente e `500` para erro inesperado.

## Variáveis de ambiente

| Variável | Padrão (desenvolvimento) |
|---|---|
| `PORT` | `8083` |
| `DATABASE_URL` | `postgres://especialidades:especialidades@localhost:5435/especialidades_db?sslmode=disable` |
| `CORS_ORIGINS` | `http://localhost:5173` (lista separada por vírgula) |

Com `DATABASE_URL=memoria` o serviço roda **sem banco**: os dados ficam na
memória e voltam ao padrão quando ele reinicia. Serve para testar o frontend
ou a triagem sem configurar o Supabase.

### Banco no Supabase

1. Crie um projeto no Supabase chamado `especialidades`.
2. Em **Connect > Session pooler**, copie a URL e troque `[YOUR-PASSWORD]`
   pela senha do banco.
3. Use essa URL em `DATABASE_URL` (no `.env` da raiz, a variável é
   `ESPECIALIDADES_DATABASE_URL`).

As tabelas e as especialidades iniciais são criadas sozinhas na primeira subida.

## Estrutura (arquitetura hexagonal, igual aos serviços em Go)

```
app/
  __main__.py            ponto de entrada: python -m app
  main.py                monta as dependências (Postgres ou memória)
  config.py              variáveis de ambiente
  ports.py               EspecialidadeRepository e VerificadorBanco
  domain/                Especialidade, DocumentoEsperado, cálculo de exames
                         faltantes, especialidades iniciais e erros
  servico/               casos de uso: listar, obter, atualizar configuração,
                         exames faltantes e cadastro inicial
  adapters/
    httpapi/             API REST: rotas, JSON em camelCase, erros, logs, CORS
    postgres/            repositório e migrações (migrations/*.sql)
    memoria/             repositório em memória (testes e DATABASE_URL=memoria)
tests/                   pytest
```

## Rodando

Pré-requisito: [Python](https://www.python.org/downloads/) 3.12+.

Na pasta `services/especialidades-service`, na primeira vez:

```powershell
py -m venv .venv
.venv\Scripts\activate
pip install -r requirements-dev.txt
```

Depois (com o `.venv` ativado):

```powershell
$env:DATABASE_URL="memoria"; python -m app     # sem banco
python -m app                                  # com o banco de DATABASE_URL
pytest                                         # testes, sem banco
```

Abra http://localhost:8083/docs para testar as rotas pelo navegador.

Para testar também o repositório Postgres, aponte `TEST_DATABASE_URL` para um
banco **descartável** (as tabelas são apagadas no teste) e rode `pytest`.

No Linux/macOS, troque `py` por `python3`, `.venv\Scripts\activate` por
`source .venv/bin/activate` e `$env:DATABASE_URL="memoria";` por
`DATABASE_URL=memoria`.
