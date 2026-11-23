import { useEffect, useState } from "react";
import { buscarHistorico, listarDocumentosDaConsulta, urlArquivo } from "../api.js";
import { dataHora, idObrigatorio, nomesTipo, tamanho } from "../utils.js";
import Cartao from "./Cartao.jsx";

/**
 * Cartão "Prontuário": histórico do paciente e documentos da consulta.
 * Recarrega o histórico sempre que `recarregar` muda (depois de um envio ou laudo).
 */
export default function Prontuario({ ids, recarregar }) {
  // null = ainda não carregou nada (lista escondida, como na versão original)
  const [lista, setLista] = useState(null);

  async function verHistorico() {
    try {
      const itens = await buscarHistorico(idObrigatorio(ids.paciente, "paciente"));
      setLista({ itens, vazio: "Nada no prontuário ainda. Envie um documento ou emita um laudo." });
    } catch (e) {
      setLista({ itens: [], vazio: e.message });
    }
  }

  async function verDaConsulta() {
    try {
      const docs = await listarDocumentosDaConsulta(idObrigatorio(ids.consulta, "consulta"));
      const itens = docs.map((d) => ({ data: d.enviadoEm, documento: { ...d, documentoId: d.id } }));
      setLista({ itens, vazio: "Nenhum documento nesta consulta." });
    } catch (e) {
      setLista({ itens: [], vazio: e.message });
    }
  }

  useEffect(() => { if (recarregar > 0) verHistorico(); }, [recarregar]); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <Cartao id="historico" largo titulo="Prontuário"
            subtitulo="Histórico clínico do paciente: documentos anexados e laudos, do mais recente ao mais antigo.">
      <div className="botoes" style={{ marginTop: 0 }}>
        <button onClick={verHistorico}>Ver histórico</button>
        <button className="secundario" onClick={verDaConsulta}>Documentos da consulta</button>
      </div>
      <div id="listaHistorico">
        {lista && (lista.itens.length === 0
          ? <div className="vazio">{lista.vazio}</div>
          : <ul className="hist">{lista.itens.map((it, i) => <ItemHistorico key={i} item={it} />)}</ul>)}
      </div>
    </Cartao>
  );
}

function ItemHistorico({ item }) {
  if (item.laudo) {
    return (
      <li>
        <span className="selo laudo">LAUDO</span>
        <div className="item">
          <strong>{item.laudo.titulo}</strong>
          <div className="meta">{dataHora(item.data)}</div>
          <p>{item.laudo.conteudo}</p>
        </div>
      </li>
    );
  }
  const d = item.documento;
  return (
    <li>
      <span className="selo">{nomesTipo[d.tipo] || d.tipo}</span>
      <div className="item">
        <strong>{d.nomeOriginal}</strong>
        <div className="meta">{`${dataHora(item.data)} · ${tamanho(d.tamanhoBytes)}`}</div>
        <a className="link" href={urlArquivo(d.documentoId)} target="_blank" rel="noopener">Abrir arquivo</a>
      </div>
    </li>
  );
}
