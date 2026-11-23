import re
import unicodedata

_NAO_ALFANUMERICO = re.compile(r"[^a-z0-9]+")


def normalizar(texto: str) -> str:
    """Minúsculas, sem acentos e só com letras/números separados por espaço.

    "Raio-X_Tórax (2).PDF" -> "raio x torax 2 pdf"
    """
    sem_acento = unicodedata.normalize("NFKD", texto).encode("ascii", "ignore").decode()
    return _NAO_ALFANUMERICO.sub(" ", sem_acento.lower()).strip()
