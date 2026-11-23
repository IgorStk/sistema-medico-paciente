# Sistema Médico–Paciente

Trabalho da disciplina **Projeto e Arquitetura de Sistemas** da Universidade de
Fortaleza (UNIFOR).

O sistema aproxima paciente e médico à distância: agendamento de consultas,
envio de exames antes do atendimento, recebimento de laudos sem deslocamento e,
como evolução futura, triagem assistida por IA por especialidade.

Monorepo com microsserviços em linguagens diferentes, integrados por um
frontend em React/Next.js que acessa os serviços via API Gateway. Os serviços
conversam entre si por REST (JSON) e cada um tem o seu próprio banco ou
armazenamento.

## Equipe

| Integrante | Linguagem |
|---|---|
| João Pedro Sales | PHP (Laravel) |
| Vitor Brito | Java (Spring Boot) |
| Michael David | Python (FastAPI) |
| Igor Roberto | Go |

## Serviços

| Integrante | Linguagem | Serviço | Operações | Banco / armazenamento | Porta local |
|---|---|---|---|---|---|
| João Pedro Sales | PHP (Laravel) | `auth` | login, logout, emitir e validar JWT, renovar sessão | PostgreSQL (usuários/sessões) | a definir |
| João Pedro Sales | PHP (Laravel) | `usuarios` | cadastrar paciente, cadastrar médico (CRM, especialidade), atualizar cadastro, consultar dados de contato | PostgreSQL (usuários/sessões) | a definir |
| Vitor Brito | Java (Spring Boot) | `agendamento` | agendar, cancelar e remarcar consulta, listar agenda do médico no dia, verificar disponibilidade, confirmar consulta | PostgreSQL (consultas) | a definir |
| Vitor Brito | Java (Spring Boot) | `notificacao` | notificar por e-mail, push ou WhatsApp, listar mensagens enviadas a um usuário | SMTP · Firebase Push · WhatsApp API | a definir |
| Michael David | Python (FastAPI) | `triagem` | solicitar triagem de uma consulta, gerar sugestão (resumo clínico, rascunho de conduta, exames faltantes) | Provedor de IA externo (API LLM) | a definir |
| Michael David | Python (FastAPI) | `especialidades` | listar especialidades, consultar/atualizar configuração (prompt base, documentos esperados), calcular exames faltantes | PostgreSQL (config. especialidade) | a definir |
| Igor Roberto | Go | [`documentos`](services/documentos-service) | enviar documento (exame prévio, laudo, receita, atestado), baixar documento, listar documentos da consulta | Amazon S3 (arquivos) + PostgreSQL `documentos_db` (metadados, 5433) | 8081 |
| Igor Roberto | Go | [`prontuario`](services/prontuario-service) | anexar documento ao prontuário, consultar histórico clínico do paciente, emitir laudo | PostgreSQL `prontuario_db` (5434) | 8082 |
| Equipe | React / Next.js | `frontend` | AppPaciente (entrar, agendar consulta, enviar exame prévio) e AppMedico (ver agenda, ver exames, triar, emitir laudo) | — | a definir |

O **API Gateway** recebe as chamadas do frontend, valida o JWT e encaminha
para o serviço certo.

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
- CORS configurável por serviço (variável `CORS_ORIGINS`).
- Commits sempre pela CLI, pelo VSCode ou pelo GitHub Desktop, nunca pela interface web do GitHub.
