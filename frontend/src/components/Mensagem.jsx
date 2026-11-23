/** Texto de retorno de uma ação: verde (ok) ou vermelho (erro). */
export default function Mensagem({ texto, ok }) {
  return <span className={"mensagem " + (texto ? (ok ? "ok" : "erro") : "")}>{texto}</span>;
}
