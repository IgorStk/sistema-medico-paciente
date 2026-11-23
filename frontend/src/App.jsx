import { useState } from "react";
import { DOCUMENTOS, PRONTUARIO } from "./api.js";
import { useIdentidade } from "./hooks/useIdentidade.js";
import EmitirLaudo from "./components/EmitirLaudo.jsx";
import EnviarDocumento from "./components/EnviarDocumento.jsx";
import Identidade from "./components/Identidade.jsx";
import Prontuario from "./components/Prontuario.jsx";
import StatusServicos from "./components/StatusServicos.jsx";
import UltimaResposta from "./components/UltimaResposta.jsx";

const servicos = [
  { nome: "documentos-service", url: DOCUMENTOS },
  { nome: "prontuario-service", url: PRONTUARIO },
];

export default function App() {
  const { ids, alterar, novoId } = useIdentidade();
  // Incrementado após um envio ou laudo, para o prontuário recarregar o histórico.
  const [recarregar, setRecarregar] = useState(0);
  const atualizarProntuario = () => setRecarregar((n) => n + 1);

  return (
    <>
      <nav>
        <a href="#identidade">Identidade</a>
        <a href="#documento">Documento</a>
        <a href="#laudo">Laudo</a>
        <a href="#historico">Prontuário</a>
      </nav>

      <header className="hero">
        <h1>Paciente e médico.<br />A um clique.</h1>
        <p>Painel de teste dos serviços de documentos e prontuário.</p>
        <StatusServicos servicos={servicos} />
      </header>

      <section className="cinza">
        <div className="grade">
          <Identidade ids={ids} alterar={alterar} novoId={novoId} />
          <EnviarDocumento ids={ids} aoEnviar={atualizarProntuario} />
          <EmitirLaudo ids={ids} aoEmitir={atualizarProntuario} />
          <Prontuario ids={ids} recarregar={recarregar} />
          <UltimaResposta />
        </div>
      </section>

      <footer>Projeto e Arquitetura de Sistemas · UNIFOR · ambiente de teste com dados fictícios</footer>
    </>
  );
}
