/** Cartão branco arredondado. `largo` ocupa a linha inteira da grade. */
export default function Cartao({ id, titulo, subtitulo, largo, children }) {
  return (
    <div className={largo ? "cartao largo" : "cartao"} id={id}>
      <h2>{titulo}</h2>
      <p className="sub">{subtitulo}</p>
      {children}
    </div>
  );
}
