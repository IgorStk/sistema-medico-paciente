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
  documentos-service/   Go: upload, consulta e download de documentos (anexa ao prontuário)
  prontuario-service/   Go: prontuário, histórico e laudos
scripts/rodar-supabase.ps1  roda os serviços Go sem Docker (chamado pelo npm run dev de services/)
frontend/               página única para testar os serviços Go pela tela
docker-compose*.yml     alternativa com Docker (opcional)
go.work                 workspace Go (para o VSCode/gopls enxergar os dois módulos)
```

## Como rodar os serviços em Go

Os bancos ficam no **Supabase**: um projeto por serviço (`documentos` e
`prontuario`). Os arquivos ficam no **Supabase Storage** do projeto
`documentos`. Nada de banco roda na sua máquina.

Pré-requisitos: [Go](https://go.dev/dl/) 1.27+ e [Node.js](https://nodejs.org/) (só pelo `npm run dev`).

1. Copie `.env.example` para `.env` e preencha com as URLs e chaves do Supabase.
   O `.env` não vai para o Git: peça os valores a quem cuida dos serviços em Go.
2. Abra **dois terminais**:

| Terminal | Pasta | Comando | O que sobe |
|---|---|---|---|
| 1 (API) | `services` | `npm run dev` | documentos-service em http://localhost:8081 e prontuario-service em http://localhost:8082 |
| 2 (front) | `frontend` | `npm run dev` | painel de teste em **http://localhost:5173** |

Ctrl+C para cada um. Na primeira vez, a API demora um pouco para compilar.
As tabelas e o bucket são criados automaticamente na primeira subida.

Para ver os dados, use o painel do Supabase: **Table Editor** para as tabelas e
**Storage** para os arquivos.

### Painel de teste (frontend mínimo)

[`frontend/index.html`](frontend/index.html) é uma página única,
sem build e sem dependências, para testar os serviços Go pela tela. Ela permite:

- gerar IDs fictícios de paciente, médico e consulta;
- enviar um documento (que é anexado ao prontuário);
- emitir um laudo;
- ver o histórico e os documentos de uma consulta;
- abrir os arquivos.

Ela precisa ser aberta por http://localhost:5173: abrir o arquivo direto
(`file://`) não funciona por causa do CORS.

### Alternativa com Docker (opcional)

Para quem tem o Docker Desktop, o `docker-compose.yml` sobe tudo localmente,
com Postgres e um S3 local (SeaweedFS), sem precisar do Supabase. Use
`docker compose up --build`. O `docker-compose.supabase.yml` faz o mesmo
apontando para o Supabase. O build de imagens não funciona com o repositório
dentro de uma pasta do OneDrive.

## Convenções da equipe

- IDs são UUID; datas em ISO 8601 (UTC).
- Respostas de erro no formato `{"erro": "mensagem"}`.
- CORS configurável por serviço (variável `CORS_ORIGINS`).
- Commits sempre pela CLI, pelo VSCode ou pelo GitHub Desktop, nunca pela interface web do GitHub.
