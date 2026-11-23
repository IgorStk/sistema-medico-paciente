"""Sobe o serviço: python -m app"""

import uvicorn

from app.config import carregar
from app.main import montar


def main() -> None:
    cfg = carregar()
    uvicorn.run(montar(cfg), host="0.0.0.0", port=cfg.porta, access_log=False)


if __name__ == "__main__":
    main()
