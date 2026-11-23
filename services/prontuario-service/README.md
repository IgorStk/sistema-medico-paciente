# prontuario-service (Go)

Microsserviço **consumidor**: mantém o prontuário de cada paciente (classe
`Prontuario` do diagrama, com `anexar()` e `historico()`) e emite laudos. Os
arquivos ficam no **documentos-service**, chamado via HTTP. O serviço tem seu
próprio banco, `prontuario_db` (Postgres).

## Endpoints

| Método | Rota | Descrição | Situação |
|---|---|---|---|
| GET | `/health` | Saúde do serviço, do banco e do documentos-service | pronto |
| GET | `/pacientes/{pacienteId}/prontuario` | Prontuário do paciente (criado no primeiro acesso) | fase 4 |
| POST | `/prontuarios/{id}/documentos` | Anexa um documento existente (`{"documentoId": "..."}`) | fase 4 |
| GET | `/prontuarios/{id}/historico` | Anexos com os metadados de cada documento | fase 5 |
| POST | `/prontuarios/{id}/laudos` | Médico emite laudo (multipart); envia ao documentos-service e anexa | fase 5 |

## Variáveis de ambiente

| Variável | Padrão |
|---|---|
| `PORT` | `8082` |
| `DATABASE_URL` | `postgres://prontuario:prontuario@localhost:5434/prontuario_db?sslmode=disable` |
| `CORS_ORIGINS` | `http://localhost:5173` |
| `DOCUMENTOS_SERVICE_URL` | `http://localhost:8081` |

## Estrutura (arquitetura hexagonal)

```
cmd/api/                 ponto de entrada (main)
internal/domain/         Prontuario, DocumentoAnexado, ItemHistorico
internal/ports/          ProntuarioPort, ProntuarioRepository, DocumentosGateway
internal/adapters/
  httpapi/               API REST, CORS, logs
  postgres/              conexão, migrações (migrations/*.sql)
  documentos/            cliente HTTP do documentos-service
internal/config/         leitura das variáveis de ambiente
```

## Rodando

```bash
go test ./...
go run ./cmd/api   # precisa do Postgres: docker compose up -d prontuario-db
```
