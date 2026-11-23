# especialidades-service (Python / FastAPI)

Microsserviço das **especialidades médicas**: lista as especialidades, guarda a
configuração de cada uma (prompt base da triagem e documentos esperados) e
calcula os exames que ainda faltam numa consulta. Quem usa essa configuração é
o `triagem-service`.

> **Situação:** Etapa 1 concluída (base do serviço). As rotas de
> especialidades e o Postgres entram na Etapa 2.

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| GET | `/health` | Saúde do serviço e do banco (`200` ok, `503` banco fora) |
| GET | `/docs` | Documentação interativa (Swagger) gerada pelo FastAPI |

Os erros vêm no formato da equipe, `{"erro": "mensagem"}`: `400` para dados
inválidos (inclusive JSON quebrado), `404` para recurso ou rota inexistente e
`500` para erro inesperado.

## Variáveis de ambiente

| Variável | Padrão (desenvolvimento) |
|---|---|
| `PORT` | `8083` |
| `DATABASE_URL` | `postgres://especialidades:especialidades@localhost:5435/especialidades_db?sslmode=disable` |
| `CORS_ORIGINS` | `http://localhost:5173` (lista separada por vírgula) |

## Estrutura (arquitetura hexagonal, igual aos serviços em Go)

```
app/
  __main__.py            ponto de entrada: python -m app
  main.py                monta as dependências
  config.py              variáveis de ambiente
  ports.py               interfaces (portas) que os adaptadores implementam
  domain/                regras de negócio e erros (DadosInvalidos, NaoEncontrado...)
  adapters/
    httpapi/             API REST: rotas, erros {"erro"}, logs e CORS
    memoria/             adaptadores em memória (testes)
tests/                   pytest, sem precisar de banco
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
python -m app      # http://localhost:8083/health e http://localhost:8083/docs
pytest             # testes, sem banco
```

No Linux/macOS, troque `py` por `python3` e `.venv\Scripts\activate` por
`source .venv/bin/activate`.
