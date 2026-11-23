import { useRef, useState } from "react";
import { enviarDocumento } from "../api.js";
import { idObrigatorio } from "../utils.js";
import Cartao from "./Cartao.jsx";
import Mensagem from "./Mensagem.jsx";

/** Cartão "Enviar documento". Chama aoEnviar() depois de um envio com sucesso. */
export default function EnviarDocumento({ ids, aoEnviar }) {
  const [tipo, setTipo] = useState("EXAME_PREVIO");
  const [enviadoPor, setEnviadoPor] = useState("paciente");
  const [enviando, setEnviando] = useState(false);
  const [msg, setMsg] = useState({ texto: "", ok: true });
  const arquivoRef = useRef(null);

  async function enviar() {
    try {
      const pacienteId = idObrigatorio(ids.paciente, "paciente");
      const autorId = idObrigatorio(ids[enviadoPor], enviadoPor);
      const arquivo = arquivoRef.current.files[0];
      if (!arquivo) throw new Error("Escolha um arquivo.");

      setEnviando(true);
      setMsg({ texto: "Enviando…", ok: true });
      const doc = await enviarDocumento({ tipo, pacienteId, autorId, consultaId: ids.consulta.trim(), arquivo });
      setMsg({ texto: `Enviado e anexado: ${doc.nomeOriginal}`, ok: true });
      arquivoRef.current.value = "";
      aoEnviar();
    } catch (e) {
      setMsg({ texto: e.message, ok: false });
    } finally {
      setEnviando(false);
    }
  }

  return (
    <Cartao id="documento" titulo="Enviar documento"
            subtitulo="PDF, JPEG, PNG ou WEBP. Vai para o Storage e é anexado ao prontuário.">
      <label htmlFor="tipo">Tipo</label>
      <select id="tipo" value={tipo} onChange={(e) => setTipo(e.target.value)}>
        <option value="EXAME_PREVIO">Exame prévio</option>
        <option value="LAUDO">Laudo (PDF)</option>
        <option value="RECEITA">Receita</option>
        <option value="ATESTADO">Atestado</option>
      </select>
      <label htmlFor="enviadoPor">Enviado por</label>
      <select id="enviadoPor" value={enviadoPor} onChange={(e) => setEnviadoPor(e.target.value)}>
        <option value="paciente">Paciente</option>
        <option value="medico">Médico</option>
      </select>
      <label htmlFor="arquivo">Arquivo</label>
      <input type="file" id="arquivo" ref={arquivoRef} accept=".pdf,.jpg,.jpeg,.png,.webp" />
      <div className="botoes">
        <button id="btEnviar" onClick={enviar} disabled={enviando}>Enviar</button>
        <Mensagem texto={msg.texto} ok={msg.ok} />
      </div>
    </Cartao>
  );
}
