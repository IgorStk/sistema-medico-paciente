-- Especialidades médicas e a configuração usada pela triagem.
CREATE TABLE especialidades (
    id            UUID PRIMARY KEY,
    codigo        TEXT        NOT NULL UNIQUE,
    nome          TEXT        NOT NULL,
    descricao     TEXT        NOT NULL DEFAULT '',
    prompt_base   TEXT        NOT NULL,
    ativa         BOOLEAN     NOT NULL DEFAULT TRUE,
    criado_em     TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Exames/documentos que o médico espera receber antes da consulta.
-- palavras_chave são procuradas no nome dos arquivos enviados.
CREATE TABLE documentos_esperados (
    especialidade_id UUID    NOT NULL REFERENCES especialidades (id) ON DELETE CASCADE,
    codigo           TEXT    NOT NULL,
    nome             TEXT    NOT NULL,
    obrigatorio      BOOLEAN NOT NULL DEFAULT TRUE,
    palavras_chave   TEXT[]  NOT NULL DEFAULT '{}',
    ordem            INTEGER NOT NULL,
    PRIMARY KEY (especialidade_id, codigo)
);
