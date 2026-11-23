# documentos-service (Go)

Microsserviço **provedor** de documentos clínicos: exames prévios, laudos,
receitas e atestados. Implementa as portas `DocumentoUploadPort` e
`DocumentoDownloadPort` do diagrama de classes. O conteúdo dos arquivos fica no
armazenamento (`ArmazenamentoS3`) e os metadados ficam no banco próprio
`documentos_db` (Postgres).

## Endpoints

| Método | Rota | Descrição | Situação |
|---|---|---|---|
| GET | `/health` | Saúde do serviço e do banco | pronto |
| POST | `/documentos` | Upload (multipart: `arquivo`, `tipo`, `pacienteId`, `consultaId?`, `autorId`) | fase 3 |
| GET | `/documentos?pacienteId=&consultaId=&tipo=` | Lista metadados | fase 3 |
| GET | `/documentos/{id}` | Metadados de um documento | fase 3 |
| GET | `/documentos/{id}/arquivo` | Download do arquivo | fase 3 |

`tipo` aceita `EXAME_PREVIO`, `LAUDO`, `RECEITA` ou `ATESTADO`.

## Variáveis de ambiente

| Variável | Padrão |
|---|---|
| `PORT` | `8081` |
| `DATABASE_URL` | `postgres://documentos:documentos@localhost:5433/documentos_db?sslmode=disable` |
| `CORS_ORIGINS` | `http://localhost:5173` (lista separada por vírgula; `*` libera todas) |
| `MAX_UPLOAD_MB` | `20` |

## Estrutura (arquitetura hexagonal)

```
cmd/api/                 ponto de entrada (main)
internal/domain/         Documento, TipoDocumento
internal/ports/          interfaces: portas de entrada e de saída
internal/adapters/
  httpapi/               API REST, CORS, logs
  postgres/              conexão, migrações (migrations/*.sql)
internal/config/         leitura das variáveis de ambiente
```

## Rodando

```bash
go test ./...
go run ./cmd/api   # precisa do Postgres: docker compose up -d documentos-db
```
