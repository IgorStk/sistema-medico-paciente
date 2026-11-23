# frontend (React + Vite)

Tela do sistema: conecta os microsserviços da equipe.

```bash
npm install     # só na primeira vez
npm run dev     # http://localhost:5173
```

A porta 5173 é fixa porque é a origem liberada no CORS dos serviços. Abrir o
`index.html` direto (`file://`) não funciona.

## Estrutura

```
src/
  main.jsx           ponto de entrada
  App.jsx            monta a tela
  index.css          estilos
  api.js             endereços dos serviços e todas as chamadas à API
  utils.js           formatação de datas, tamanhos etc.
  hooks/
    useIdentidade.js IDs de teste guardados no navegador (até existir login)
  components/        cartões e peças da tela
```

Para integrar um serviço novo:
1. Adicione a URL e as funções de chamada em `api.js`.
2. Crie os componentes em `components/` e use-os no `App.jsx`.
3. No serviço, libere o CORS para `http://localhost:5173`.

Os endereços podem ser trocados sem mexer no código, criando
`frontend/.env.local` (não vai para o Git):

```
VITE_DOCUMENTOS_URL=http://localhost:8081
VITE_PRONTUARIO_URL=http://localhost:8082
```
