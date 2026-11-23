"""Rotas de especialidades. `{ref}` aceita o UUID ou o código (ex.: cardiologia)."""

from fastapi import APIRouter, Path, Query, Request

from app.servico.especialidades import ServicoEspecialidades

from .esquemas import (
    AtualizarConfiguracao,
    Configuracao,
    EspecialidadeDetalhe,
    EspecialidadeResumo,
    PedidoExamesFaltantes,
    RespostaExamesFaltantes,
)

router = APIRouter(prefix="/especialidades", tags=["especialidades"])

REF = Path(description="UUID ou código da especialidade (ex.: cardiologia)")
ERROS = {400: {"description": "Dados inválidos"}, 404: {"description": "Especialidade não encontrada"}}


def _servico(request: Request) -> ServicoEspecialidades:
    return request.app.state.dep.especialidades


@router.get("", response_model=list[EspecialidadeResumo])
def listar(request: Request, todas: bool = Query(False, description="true inclui as inativas")):
    """Lista as especialidades (por padrão só as ativas), em ordem alfabética."""
    return [EspecialidadeResumo.de(e) for e in _servico(request).listar(apenas_ativas=not todas)]


@router.get("/{ref}", response_model=EspecialidadeDetalhe, responses=ERROS)
def obter(request: Request, ref: str = REF):
    """Dados da especialidade, incluindo a configuração da triagem."""
    return EspecialidadeDetalhe.de(_servico(request).obter(ref))


@router.get("/{ref}/configuracao", response_model=Configuracao, responses=ERROS)
def obter_configuracao(request: Request, ref: str = REF):
    """Prompt base e documentos esperados (usado pelo triagem-service)."""
    return Configuracao.de(_servico(request).obter(ref))


@router.put("/{ref}/configuracao", response_model=Configuracao, responses=ERROS)
def atualizar_configuracao(corpo: AtualizarConfiguracao, request: Request, ref: str = REF):
    """Substitui o prompt base e a lista inteira de documentos esperados."""
    atualizada = _servico(request).atualizar_configuracao(
        ref, corpo.prompt_base, [d.para_dominio() for d in corpo.documentos_esperados]
    )
    return Configuracao.de(atualizada)


@router.post("/{ref}/exames-faltantes", response_model=RespostaExamesFaltantes, responses=ERROS)
def exames_faltantes(corpo: PedidoExamesFaltantes, request: Request, ref: str = REF):
    """Compara os documentos já enviados com os esperados pela especialidade."""
    especialidade, resultado = _servico(request).exames_faltantes(ref, corpo.documentos)
    return RespostaExamesFaltantes.de(especialidade, resultado)
