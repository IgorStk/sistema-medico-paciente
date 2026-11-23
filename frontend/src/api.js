// Todas as chamadas aos microsserviços.
// Os endereços podem ser trocados em frontend/.env.local (ex.: VITE_DOCUMENTOS_URL=...).
export const DOCUMENTOS = import.meta.env.VITE_DOCUMENTOS_URL || "http://localhost:8081";
export const PRONTUARIO = import.meta.env.VITE_PRONTUARIO_URL || "http://localhost:8082";

// ---------- Base ----------

// Guarda a última resposta para o cartão "Última resposta".
let ultima = "—";
const ouvintes = new Set();

export const ultimaResposta = {
  assinar(avisar) {
    ouvintes.add(avisar);
    return () => ouvintes.delete(avisar);
  },
  ler: () => ultima,
};

function registrar(metodo, url, status, corpo) {
  const texto = typeof corpo === "string" ? corpo : JSON.stringify(corpo, null, 2);
  ultima = `${metodo} ${url}\n→ ${status}\n\n${texto}`;
  ouvintes.forEach((avisar) => avisar());
}

/** Faz a requisição e devolve o corpo. Lança Error com a mensagem da API se o status não for 2xx. */
async function chamar(metodo, url, opcoes = {}) {
  let resp;
  try {
    resp = await fetch(url, { method: metodo, ...opcoes });
  } catch {
    registrar(metodo, url, "falha de conexão", "Serviço fora do ar? Rode npm run dev na pasta services.");
    throw new Error("Não consegui falar com o serviço.");
  }
  const texto = await resp.text();
  let corpo = texto;
  try { corpo = JSON.parse(texto); } catch { /* resposta não é JSON */ }
  registrar(metodo, url, resp.status, corpo);
  if (!resp.ok) throw new Error(corpo?.erro || `Erro ${resp.status}`);
  return corpo;
}

export async function saudavel(base) {
  try {
    return (await fetch(base + "/health")).ok;
  } catch {
    return false;
  }
}

const json = (corpo) => ({ headers: { "Content-Type": "application/json" }, body: JSON.stringify(corpo) });

// ---------- documentos-service ----------

export function enviarDocumento({ tipo, pacienteId, autorId, consultaId, arquivo }) {
  const form = new FormData();
  form.append("tipo", tipo);
  form.append("pacienteId", pacienteId);
  form.append("autorId", autorId);
  if (consultaId) form.append("consultaId", consultaId);
  form.append("arquivo", arquivo);
  return chamar("POST", `${DOCUMENTOS}/documentos`, { body: form });
}

export function listarDocumentosDaConsulta(consultaId) {
  return chamar("GET", `${DOCUMENTOS}/documentos?consultaId=${encodeURIComponent(consultaId)}`);
}

export function urlArquivo(documentoId) {
  return `${DOCUMENTOS}/documentos/${documentoId}/arquivo?inline=true`;
}

// ---------- prontuario-service ----------

const prontuarioDe = (pacienteId) => `${PRONTUARIO}/pacientes/${encodeURIComponent(pacienteId)}/prontuario`;

export function buscarHistorico(pacienteId) {
  return chamar("GET", `${prontuarioDe(pacienteId)}/historico`);
}

export function emitirLaudo(pacienteId, { medicoId, titulo, conteudo, consultaId }) {
  const corpo = { medicoId, titulo, conteudo };
  if (consultaId) corpo.consultaId = consultaId;
  return chamar("POST", `${prontuarioDe(pacienteId)}/laudos`, json(corpo));
}
