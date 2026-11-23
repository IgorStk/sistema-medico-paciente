# documentos-service (Go)

Microsserviço **provedor** de documentos clínicos: exames prévios, laudos,
receitas e atestados. Implementa as portas `DocumentoUploadPort` e
`DocumentoDownloadPort` do diagrama de classes.

- **Arquivos:** ficam no S3 (`ArmazenamentoS3`, com o SDK oficial da AWS). Em produção é o Supabase Storage, pela API compatível com S3; no docker-compose local é o SeaweedFS. O Amazon S3 também funciona, só mudando a configuração.
- **Metadados:** ficam no banco próprio `documentos_db` (Postgres).
- **Consumidor do prontuário:** depois de cada upload, o serviço chama o
  prontuario-service para anexar o documento ao prontuário do paciente (seta
  "anexar" do diagrama de microsserviços).

### Como funciona o upload

1. O arquivo é gravado no S3.
2. Os metadados são gravados no `documentos_db`.
3. O documento é anexado ao prontuário (`POST /pacientes/{pacienteId}/prontuario/documentos` no prontuario-service).

Se um passo falha, os anteriores são desfeitos. Assim, um documento só existe
se estiver no prontuário, e não sobra arquivo órfão no S3. Se o prontuário
estiver fora do ar, o upload responde `502` e pode ser repetido.

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| GET | `/health` | Saúde do serviço, do banco e do S3 (e, em `dependencias`, do prontuario-service) |
| POST | `/documentos` | Upload (multipart, campos abaixo) → `201` com o documento |
| GET | `/documentos?pacienteId=&consultaId=&tipo=` | Lista metadados (exige `pacienteId` ou `consultaId`) |
| GET | `/documentos/{id}` | Metadados de um documento |
| GET | `/documentos/{id}/arquivo` | Download do arquivo (`?inline=true` para abrir no navegador) |

### Upload: campos do `multipart/form-data`

| Campo | Obrigatório | Descrição |
|---|---|---|
| `arquivo` | sim | PDF, JPEG, PNG ou WEBP, até `MAX_UPLOAD_MB` |
| `tipo` | sim | `EXAME_PREVIO`, `LAUDO`, `RECEITA` ou `ATESTADO` |
| `pacienteId` | sim | UUID do paciente |
| `autorId` | sim | UUID de quem envia (depois virá do JWT) |
| `consultaId` | não | UUID da consulta (serviço de agendamento) |

O formato é detectado pelo **conteúdo** do arquivo, não pela extensão.

Exemplo:

```bash
curl -F tipo=EXAME_PREVIO -F pacienteId=<uuid> -F autorId=<uuid> \
     -F consultaId=<uuid> -F arquivo=@hemograma.pdf http://localhost:8081/documentos
```

Resposta `201`:

```json
{
  "id": "503f07c9-72bc-4a39-8904-65e08d8f6f43",
  "tipo": "EXAME_PREVIO",
  "pacienteId": "…",
  "consultaId": "…",
  "autorId": "…",
  "nomeOriginal": "hemograma.pdf",
  "contentType": "application/pdf",
  "tamanhoBytes": 48213,
  "enviadoEm": "2026-11-23T16:44:57.738563Z"
}
```

### Erros

Os erros vêm sempre no formato `{"erro": "mensagem"}`.

| Status | Quando |
|---|---|
| 400 | campo ausente ou inválido, tipo desconhecido, listagem sem filtro |
| 404 | documento não existe |
| 413 | arquivo maior que o limite |
| 415 | formato não permitido |
| 502 | o prontuario-service não confirmou o anexo (o upload foi desfeito) |

## Variáveis de ambiente

| Variável | Padrão (desenvolvimento) |
|---|---|
| `PORT` | `8081` |
| `DATABASE_URL` | `postgres://documentos:documentos@localhost:5433/documentos_db?sslmode=disable` |
| `CORS_ORIGINS` | `http://localhost:5173` (lista separada por vírgula; `*` libera todas) |
| `MAX_UPLOAD_MB` | `20` |
| `S3_ENDPOINT` | `http://localhost:8333` (URL completa; Supabase: `https://<ref>.storage.supabase.co/storage/v1/s3`) |
| `S3_ACCESS_KEY` / `S3_SECRET_KEY` | credenciais de desenvolvimento do `dev/seaweedfs-s3.json` |
| `S3_BUCKET` | `documentos` (criado automaticamente na subida) |
| `S3_REGION` | `us-east-1` |
| `PRONTUARIO_SERVICE_URL` | `http://localhost:8082` |

## Estrutura (arquitetura hexagonal)

```
cmd/api/                 ponto de entrada: monta as dependências
internal/domain/         Documento, TipoDocumento, erros de domínio
internal/ports/          portas de entrada (Upload/Download) e de saída (repositório, armazenamento)
internal/servico/        casos de uso: validação, upload com compensação, download
internal/adapters/
  httpapi/               API REST, CORS, logs
  postgres/              repositório e migrações (migrations/*.sql)
  s3/                    ArmazenamentoS3
  prontuario/            cliente HTTP do prontuario-service (anexar)
  memoria/               repositório, armazenamento e prontuário em memória (testes)
internal/config/         variáveis de ambiente
dev/                     configuração do S3 local
```

## Rodando

```bash
go test ./...                     # não precisa de banco nem S3
docker compose up --build -d      # na raiz do repositório
```
