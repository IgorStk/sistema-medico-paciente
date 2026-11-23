-- Um prontuário por paciente. paciente_id vem do serviço de usuários,
-- por isso não há chave estrangeira (cada microsserviço tem o seu banco).
CREATE TABLE prontuarios (
    id          UUID PRIMARY KEY,
    paciente_id UUID        NOT NULL UNIQUE,
    criado_em   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Documentos anexados ao prontuário. documento_id aponta para um registro
-- do documentos-service (outro banco), então também não tem chave estrangeira.
CREATE TABLE prontuario_documentos (
    prontuario_id UUID        NOT NULL REFERENCES prontuarios (id) ON DELETE CASCADE,
    documento_id  UUID        NOT NULL,
    anexado_por   UUID        NOT NULL,
    anexado_em    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (prontuario_id, documento_id)
);

CREATE INDEX idx_prontuario_documentos_data ON prontuario_documentos (prontuario_id, anexado_em DESC);
