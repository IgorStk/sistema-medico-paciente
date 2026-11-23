import uuid

from psycopg import Connection
from psycopg_pool import ConnectionPool

from app.domain.especialidade import DocumentoEsperado, Especialidade

_COLUNAS = "id, codigo, nome, descricao, prompt_base, ativa, atualizado_em"


class EspecialidadeRepositoryPostgres:
    def __init__(self, pool: ConnectionPool):
        self.pool = pool

    def listar(self, apenas_ativas: bool) -> list[Especialidade]:
        filtro = "WHERE ativa" if apenas_ativas else ""
        with self.pool.connection() as conn:
            linhas = conn.execute(f"SELECT {_COLUNAS} FROM especialidades {filtro} ORDER BY nome").fetchall()
            return [self._montar(conn, linha) for linha in linhas]

    def buscar_por_id(self, id: uuid.UUID) -> Especialidade | None:
        return self._buscar("id = %s", id)

    def buscar_por_codigo(self, codigo: str) -> Especialidade | None:
        return self._buscar("codigo = %s", codigo)

    def salvar_configuracao(self, e: Especialidade) -> None:
        with self.pool.connection() as conn, conn.transaction():
            conn.execute(
                "UPDATE especialidades SET prompt_base = %s, atualizado_em = %s WHERE id = %s",
                (e.prompt_base, e.atualizado_em, e.id),
            )
            conn.execute("DELETE FROM documentos_esperados WHERE especialidade_id = %s", (e.id,))
            self._inserir_documentos(conn, e)

    def inserir_se_nao_existir(self, e: Especialidade) -> bool:
        with self.pool.connection() as conn, conn.transaction():
            inserida = conn.execute(
                """INSERT INTO especialidades (id, codigo, nome, descricao, prompt_base, ativa, atualizado_em)
                   VALUES (%s, %s, %s, %s, %s, %s, %s)
                   ON CONFLICT DO NOTHING
                   RETURNING id""",
                (e.id, e.codigo, e.nome, e.descricao, e.prompt_base, e.ativa, e.atualizado_em),
            ).fetchone()
            if inserida:
                self._inserir_documentos(conn, e)
            return inserida is not None

    # ---------- auxiliares ----------

    def _buscar(self, condicao: str, valor) -> Especialidade | None:
        with self.pool.connection() as conn:
            linha = conn.execute(f"SELECT {_COLUNAS} FROM especialidades WHERE {condicao}", (valor,)).fetchone()
            return self._montar(conn, linha) if linha else None

    @staticmethod
    def _inserir_documentos(conn: Connection, e: Especialidade) -> None:
        for ordem, d in enumerate(e.documentos_esperados):
            conn.execute(
                """INSERT INTO documentos_esperados
                   (especialidade_id, codigo, nome, obrigatorio, palavras_chave, ordem)
                   VALUES (%s, %s, %s, %s, %s, %s)""",
                (e.id, d.codigo, d.nome, d.obrigatorio, list(d.palavras_chave), ordem),
            )

    @staticmethod
    def _montar(conn: Connection, linha) -> Especialidade:
        id_, codigo, nome, descricao, prompt, ativa, atualizado_em = linha
        docs = conn.execute(
            """SELECT codigo, nome, obrigatorio, palavras_chave FROM documentos_esperados
               WHERE especialidade_id = %s ORDER BY ordem""",
            (id_,),
        ).fetchall()
        return Especialidade(
            id=id_,
            codigo=codigo,
            nome=nome,
            descricao=descricao,
            prompt_base=prompt,
            documentos_esperados=tuple(DocumentoEsperado(c, n, o, tuple(p)) for c, n, o, p in docs),
            ativa=ativa,
            atualizado_em=atualizado_em,
        )
