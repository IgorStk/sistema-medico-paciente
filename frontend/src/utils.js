export const dataHora = (iso) => new Date(iso).toLocaleString("pt-BR", { dateStyle: "short", timeStyle: "short" });

export const tamanho = (b) =>
  b < 1024 ? `${b} B` : b < 1048576 ? `${(b / 1024).toFixed(1)} KB` : `${(b / 1048576).toFixed(1)} MB`;

export const nomesTipo = { EXAME_PREVIO: "Exame prévio", LAUDO: "Laudo", RECEITA: "Receita", ATESTADO: "Atestado" };

/** Devolve o ID preenchido ou lança um erro com mensagem amigável. */
export function idObrigatorio(valor, nome) {
  const v = (valor || "").trim();
  if (!v) throw new Error(`Preencha o ID do ${nome}.`);
  return v;
}
