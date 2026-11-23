-- Seguindo o diagrama de microsserviços, é o documentos-service que chama o
-- prontuário ("anexar"). O anexo passa a guardar uma cópia dos metadados do
-- documento, para o histórico não depender do outro serviço.
-- (A tabela estava vazia: o prontuário ainda não tinha rotas de escrita.)
ALTER TABLE prontuario_documentos
    DROP COLUMN anexado_por,
    ADD COLUMN tipo          TEXT        NOT NULL CHECK (tipo IN ('EXAME_PREVIO', 'LAUDO', 'RECEITA', 'ATESTADO')),
    ADD COLUMN nome_original TEXT        NOT NULL,
    ADD COLUMN content_type  TEXT        NOT NULL,
    ADD COLUMN tamanho_bytes BIGINT      NOT NULL CHECK (tamanho_bytes >= 0),
    ADD COLUMN consulta_id   UUID,
    ADD COLUMN autor_id      UUID        NOT NULL,
    ADD COLUMN enviado_em    TIMESTAMPTZ NOT NULL;

-- Laudos escritos pelo médico direto no prontuário ("emitir laudo").
CREATE TABLE laudos (
    id            UUID PRIMARY KEY,
    prontuario_id UUID        NOT NULL REFERENCES prontuarios (id) ON DELETE CASCADE,
    consulta_id   UUID,
    medico_id     UUID        NOT NULL,
    titulo        TEXT        NOT NULL,
    conteudo      TEXT        NOT NULL,
    emitido_em    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_laudos_prontuario ON laudos (prontuario_id, emitido_em DESC);
