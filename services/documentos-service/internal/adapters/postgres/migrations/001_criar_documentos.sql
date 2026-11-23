-- Metadados dos documentos. O conteúdo binário fica no armazenamento de
-- arquivos (ArmazenamentoS3), referenciado por chave_armazenamento.
-- paciente_id, consulta_id e autor_id pertencem a outros serviços, por isso
-- não há chave estrangeira: cada microsserviço tem o seu banco.
CREATE TABLE documentos (
    id                  UUID PRIMARY KEY,
    tipo                TEXT        NOT NULL CHECK (tipo IN ('EXAME_PREVIO', 'LAUDO', 'RECEITA', 'ATESTADO')),
    paciente_id         UUID        NOT NULL,
    consulta_id         UUID,
    autor_id            UUID        NOT NULL,
    nome_original       TEXT        NOT NULL,
    content_type        TEXT        NOT NULL,
    tamanho_bytes       BIGINT      NOT NULL CHECK (tamanho_bytes >= 0),
    chave_armazenamento TEXT        NOT NULL UNIQUE,
    enviado_em          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_documentos_paciente ON documentos (paciente_id, enviado_em DESC);
CREATE INDEX idx_documentos_consulta ON documentos (consulta_id) WHERE consulta_id IS NOT NULL;
