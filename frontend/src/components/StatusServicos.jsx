import { useEffect, useState } from "react";
import { saudavel } from "../api.js";

/** Pontinhos de saúde (verde/vermelho) para cada serviço. */
export default function StatusServicos({ servicos }) {
  const [estados, setEstados] = useState({});

  async function verificar() {
    setEstados({});
    for (const { nome, url } of servicos) {
      const ok = await saudavel(url);
      setEstados((atual) => ({ ...atual, [nome]: ok ? "ok" : "erro" }));
    }
  }

  useEffect(() => { verificar(); }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="status">
      {servicos.map(({ nome }) => (
        <span key={nome}><span className={"ponto " + (estados[nome] || "")}></span>{nome}</span>
      ))}
      <button className="mini" onClick={verificar}>Verificar de novo</button>
    </div>
  );
}
