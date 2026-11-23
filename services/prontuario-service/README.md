# prontuario-service (Go)

Microsserviço **provedor** do prontuário de cada paciente (classe `Prontuario`
do diagrama, com `anexar()` e `historico()`), mais a emissão de laudos. Tem
banco próprio: `prontuario_db` (Postgres, no Supabase em produção).

Seguindo o diagrama de microsserviços, quem chama o prontuário é o
**documentos-service** (seta "anexar"). Depois de cada upload, ele envia os
metadados do documento para cá. O prontuário guarda uma cópia desses
metadados, então o histórico não depende do outro serviço. O arquivo em si
continua no documentos-service (`GET /documentos/{id}/arquivo`).

## Endpoints

Tudo é endereçado pelo paciente. O prontuário é criado no primeiro acesso.

| Método | Rota | Quem chama | Descrição |
|---|---|---|---|
| GET | `/health` | — | Saúde do serviço e do banco |
| GET | `/pacientes/{pacienteId}/prontuario` | frontend | Dados do prontuário |
| GET | `/pacientes/{pacienteId}/prontuario/historico` | frontend | Documentos e laudos, do mais recente ao mais antigo |
| POST | `/pacientes/{pacienteId}/prontuario/documentos` | documentos-service | Anexa um documento (`201`; `200` se já estava anexado) |
| POST | `/pacientes/{pacienteId}/prontuario/laudos` | frontend (AppMedico) | Médico emite um laudo escrito → `201` |

### Emitir laudo

```json
POST /pacientes/{pacienteId}/prontuario/laudos
{
  "medicoId": "uuid",
  "consultaId": "uuid (opcional)",
  "titulo": "Laudo do hemograma",
  "conteudo": "Hemograma sem alterações."
}
```

Os laudos em PDF seguem outro caminho: são enviados ao documentos-service com
`tipo=LAUDO` e chegam ao prontuário como documento anexado.

### Histórico

```json
[
  { "tipo": "LAUDO", "data": "2026-11-23T17:56:07Z", "laudo": { "id": "…", "titulo": "…", "conteudo": "…" } },
  { "tipo": "DOCUMENTO", "data": "2026-11-23T17:56:05Z", "documento": { "documentoId": "…", "tipo": "EXAME_PREVIO", "nomeOriginal": "hemograma.pdf" } }
]
```

### Anexar documento (contrato com o documentos-service)

```json
POST /pacientes/{pacienteId}/prontuario/documentos
{
  "documentoId": "uuid",
  "tipo": "EXAME_PREVIO | LAUDO | RECEITA | ATESTADO",
  "nomeOriginal": "hemograma.pdf",
  "contentType": "application/pdf",
  "tamanhoBytes": 48213,
  "consultaId": "uuid (opcional)",
  "autorId": "uuid",
  "enviadoEm": "2026-11-23T17:56:05Z"
}
```

Os erros vêm no formato `{"erro": "mensagem"}`: `400` para dados inválidos e
`413` para corpo grande demais.

## Variáveis de ambiente

| Variável | Padrão (desenvolvimento) |
|---|---|
| `PORT` | `8082` |
| `DATABASE_URL` | `postgres://prontuario:prontuario@localhost:5434/prontuario_db?sslmode=disable` |
| `CORS_ORIGINS` | `http://localhost:5173` |

## Estrutura (arquitetura hexagonal)

```
cmd/api/                 ponto de entrada: monta as dependências
internal/domain/         Prontuario, DocumentoAnexado, Laudo, ItemHistorico
internal/ports/          ProntuarioPort (entrada) e ProntuarioRepository (saída)
internal/servico/        casos de uso: obter/criar, anexar (idempotente), laudo, histórico
internal/adapters/
  httpapi/               API REST, CORS, logs
  postgres/              repositório e migrações (migrations/*.sql)
  memoria/               repositório em memória (testes)
internal/config/         variáveis de ambiente
```

## Rodando

```bash
go test ./...                     # não precisa de banco
.\scripts\rodar-supabase.ps1      # na raiz: sobe este serviço e o documentos-service
```
