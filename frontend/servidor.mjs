// Servidor estático mínimo para o painel de teste (sem dependências).
// Uso: npm run dev  →  http://localhost:5173
import { createServer } from "node:http";
import { readFile } from "node:fs/promises";
import { extname, join, normalize } from "node:path";
import { fileURLToPath } from "node:url";

const PORTA = 5173; // liberada no CORS dos serviços Go
const PASTA = fileURLToPath(new URL(".", import.meta.url));
const TIPOS = { ".html": "text/html; charset=utf-8", ".js": "text/javascript", ".mjs": "text/javascript", ".css": "text/css", ".svg": "image/svg+xml", ".png": "image/png", ".ico": "image/x-icon" };

createServer(async (req, res) => {
  const caminho = decodeURIComponent(new URL(req.url, "http://localhost").pathname);
  const arquivo = normalize(join(PASTA, caminho === "/" ? "index.html" : caminho));
  if (!arquivo.startsWith(PASTA)) { res.writeHead(403).end(); return; } // impede ../
  try {
    const conteudo = await readFile(arquivo);
    res.writeHead(200, { "Content-Type": TIPOS[extname(arquivo)] || "application/octet-stream", "Cache-Control": "no-store" });
    res.end(conteudo);
  } catch {
    res.writeHead(404, { "Content-Type": "text/plain; charset=utf-8" }).end("Não encontrado");
  }
}).listen(PORTA, "localhost", () => {
  console.log(`Painel de teste em http://localhost:${PORTA}  (Ctrl+C para parar)`);
});
