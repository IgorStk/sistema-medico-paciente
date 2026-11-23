import { useSyncExternalStore } from "react";
import { ultimaResposta } from "../api.js";
import Cartao from "./Cartao.jsx";

/** Mostra a última resposta de qualquer chamada feita pelo api.js. */
export default function UltimaResposta() {
  const texto = useSyncExternalStore(ultimaResposta.assinar, ultimaResposta.ler);
  return (
    <Cartao titulo="Última resposta" subtitulo="O que a API devolveu na última chamada." largo>
      <pre className="log">{texto}</pre>
    </Cartao>
  );
}
