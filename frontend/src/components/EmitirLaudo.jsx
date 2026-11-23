import { useState } from "react";
import { emitirLaudo } from "../api.js";
import { idObrigatorio } from "../utils.js";
import Cartao from "./Cartao.jsx";
import Mensagem from "./Mensagem.jsx";

/** Cartão "Emitir laudo". Chama aoEmitir() depois de um laudo emitido. */
export default function EmitirLaudo({ ids, aoEmitir }) {
  const [titulo, setTitulo] = useState("Laudo do hemograma");
  const [conteudo, setConteudo] = useState("Hemograma sem alterações significativas.");
  const [enviando, setEnviando] = useState(false);
  const [msg, setMsg] = useState({ texto: "", ok: true });

  async function emitir() {
    try {
      const pacienteId = idObrigatorio(ids.paciente, "paciente");
      const medicoId = idObrigatorio(ids.medico, "médico");
      setEnviando(true);
      await emitirLaudo(pacienteId, { medicoId, titulo, conteudo, consultaId: ids.consulta.trim() });
      setMsg({ texto: "Laudo emitido.", ok: true });
      aoEmitir();
    } catch (e) {
      setMsg({ texto: e.message, ok: false });
    } finally {
      setEnviando(false);
    }
  }

  return (
    <Cartao id="laudo" titulo="Emitir laudo" subtitulo="O médico escreve o laudo direto no prontuário do paciente.">
      <label htmlFor="titulo">Título</label>
      <input type="text" id="titulo" value={titulo} onChange={(e) => setTitulo(e.target.value)}
             style={{ fontFamily: "inherit", fontSize: 15 }} />
      <label htmlFor="conteudo">Conteúdo</label>
      <textarea id="conteudo" value={conteudo} onChange={(e) => setConteudo(e.target.value)} />
      <div className="botoes">
        <button id="btLaudo" onClick={emitir} disabled={enviando}>Emitir laudo</button>
        <Mensagem texto={msg.texto} ok={msg.ok} />
      </div>
    </Cartao>
  );
}
