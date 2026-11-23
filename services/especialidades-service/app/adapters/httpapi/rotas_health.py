from fastapi import APIRouter, Request
from fastapi.responses import JSONResponse
from pydantic import BaseModel

router = APIRouter(tags=["saúde"])


class StatusResposta(BaseModel):
    status: str
    servico: str
    banco: str


@router.get("/health", response_model=StatusResposta, responses={503: {"model": StatusResposta}})
def health(request: Request):
    """Saúde do serviço e do banco: 200 se está tudo ok, 503 se o banco não responde."""
    dep = request.app.state.dep
    try:
        dep.banco.ping()
    except Exception as erro:  # qualquer falha no ping conta como banco fora
        dep.log.warning("banco indisponível no /health: %s", erro)
        return JSONResponse(
            status_code=503,
            content=StatusResposta(status="degradado", servico=dep.servico, banco="indisponivel").model_dump(),
        )
    return StatusResposta(status="ok", servico=dep.servico, banco="ok")
