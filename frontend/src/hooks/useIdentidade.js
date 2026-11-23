// IDs fictícios de paciente, médico e consulta, guardados no navegador.
// Quando o login existir, isto dá lugar ao usuário logado.
import { useState } from "react";

function ler(chave) { try { return localStorage.getItem(chave); } catch { return null; } }
function gravar(chave, valor) { try { localStorage.setItem(chave, valor); } catch { /* sem storage */ } }

function inicial(campo) {
  const salvo = ler(campo);
  if (salvo !== null) return salvo;
  if (campo === "consulta") return "";
  const novo = crypto.randomUUID();
  gravar(campo, novo);
  return novo;
}

export function useIdentidade() {
  const [ids, setIds] = useState(() => ({
    paciente: inicial("paciente"),
    medico: inicial("medico"),
    consulta: inicial("consulta"),
  }));

  function alterar(campo, valor) {
    gravar(campo, valor.trim());
    setIds((atual) => ({ ...atual, [campo]: valor }));
  }

  return { ids, alterar, novoId: (campo) => alterar(campo, crypto.randomUUID()) };
}
