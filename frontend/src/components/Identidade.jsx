import Cartao from "./Cartao.jsx";

/** Cartão "Identidade de teste": IDs fictícios de paciente, médico e consulta. */
export default function Identidade({ ids, alterar, novoId }) {
  const campo = (rotulo, nome, placeholder) => (
    <>
      <label htmlFor={nome}>{rotulo}</label>
      <div className="linha">
        <input type="text" id={nome} value={ids[nome]} placeholder={placeholder}
               onChange={(e) => alterar(nome, e.target.value)} />
        <button className="mini" onClick={() => novoId(nome)}>Novo</button>
      </div>
    </>
  );

  return (
    <Cartao id="identidade" titulo="Identidade de teste"
            subtitulo="Ainda não há login. Use IDs fictícios; eles ficam salvos neste navegador.">
      {campo("Paciente", "paciente")}
      {campo("Médico", "medico")}
      {campo("Consulta (opcional)", "consulta", "vazio = sem consulta")}
    </Cartao>
  );
}
