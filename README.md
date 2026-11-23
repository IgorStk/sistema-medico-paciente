# Sistema Médico–Paciente

Trabalho da disciplina **Projeto e Arquitetura de Sistemas** da Universidade de
Fortaleza (UNIFOR).

O sistema aproxima paciente e médico à distância: agendamento de consultas,
envio de exames antes do atendimento, recebimento de laudos sem deslocamento e,
como evolução futura, triagem assistida por IA por especialidade.

Monorepo com microsserviços em linguagens diferentes, integrados por um
frontend em React. Cada serviço tem o seu próprio banco de dados.

## Equipe e serviços

| Integrante | Linguagem | Serviço | Porta | Banco |
|---|---|---|---|---|
| _(seu nome)_ | Go | `documentos-service` (provedor) | 8081 | `documentos_db` (Postgres, 5433) |
| _(seu nome)_ | Go | `prontuario-service` (consumidor) | 8082 | `prontuario_db` (Postgres, 5434) |
| _(integrante)_ | _(linguagem)_ | _(serviço)_ | | |
| _(integrante)_ | _(linguagem)_ | _(serviço)_ | | |
| _(equipe)_ | React | `frontend` | 5173 | — |

## Estrutura

```
services/
  documentos-service/   Go: upload, consulta e download de documentos
  prontuario-service/   Go: prontuário, histórico e laudos (consome documentos-service)
docker-compose.yml      sobe todos os serviços e bancos
go.work                 workspace Go (para o VSCode/gopls enxergar os dois módulos)
```

## Como rodar

Pré-requisito: [Docker Desktop](https://www.docker.com/products/docker-desktop/).

```bash
docker compose up --build
```

Teste rápido:

```bash
curl http://localhost:8081/health
curl http://localhost:8082/health
```

## Convenções da equipe

- IDs são UUID; datas em ISO 8601 (UTC).
- Respostas de erro no formato `{"erro": "mensagem"}`.
- CORS liberado para o frontend em `http://localhost:5173`.
- Commits sempre pela CLI, pelo VSCode ou pelo GitHub Desktop, nunca pela interface web do GitHub.
