// Package servico implementa os casos de uso (portas de entrada
// DocumentoUploadPort e DocumentoDownloadPort) usando as portas de saída.
package servico

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"

	"documentos-service/internal/domain"
	"documentos-service/internal/ports"
)

const tamanhoMaxNome = 255

// Documentos implementa ports.DocumentoUploadPort e ports.DocumentoDownloadPort.
type Documentos struct {
	repo          ports.DocumentoRepository
	armazenamento ports.ArmazenamentoArquivos
	tamanhoMax    int64
	agora         func() time.Time
}

var (
	_ ports.DocumentoUploadPort   = (*Documentos)(nil)
	_ ports.DocumentoDownloadPort = (*Documentos)(nil)
)

// NovoDocumentos cria o serviço; tamanhoMaxBytes limita cada arquivo.
func NovoDocumentos(repo ports.DocumentoRepository, armazenamento ports.ArmazenamentoArquivos, tamanhoMaxBytes int64) *Documentos {
	return &Documentos{repo: repo, armazenamento: armazenamento, tamanhoMax: tamanhoMaxBytes, agora: time.Now}
}

// Enviar grava o arquivo no armazenamento e depois os metadados no banco. Se
// o banco falhar, remove o arquivo para não deixar objeto órfão no S3.
func (s *Documentos) Enviar(ctx context.Context, novo domain.NovoDocumento, conteudo io.Reader) (domain.Documento, error) {
	novo.NomeOriginal = limparNome(novo.NomeOriginal)
	if err := s.validar(novo); err != nil {
		return domain.Documento{}, err
	}

	id := uuid.New()
	doc := domain.Documento{
		ID:           id,
		Tipo:         novo.Tipo,
		PacienteID:   novo.PacienteID,
		ConsultaID:   novo.ConsultaID,
		AutorID:      novo.AutorID,
		NomeOriginal: novo.NomeOriginal,
		ContentType:  novo.ContentType,
		TamanhoBytes: novo.TamanhoBytes,
		URI:          fmt.Sprintf("%s/%s", novo.PacienteID, id),
		// O Postgres guarda microssegundos; truncar mantém a resposta igual à leitura posterior.
		EnviadoEm: s.agora().UTC().Truncate(time.Microsecond),
	}

	if err := s.armazenamento.Salvar(ctx, doc.URI, conteudo, doc.TamanhoBytes, doc.ContentType); err != nil {
		return domain.Documento{}, err
	}
	if err := s.repo.Salvar(ctx, doc); err != nil {
		if errRemover := s.armazenamento.Remover(context.WithoutCancel(ctx), doc.URI); errRemover != nil {
			return domain.Documento{}, errors.Join(err, errRemover)
		}
		return domain.Documento{}, err
	}
	return doc, nil
}

func (s *Documentos) BuscarPorID(ctx context.Context, id uuid.UUID) (domain.Documento, error) {
	return s.repo.BuscarPorID(ctx, id)
}

// Listar exige ao menos pacienteId ou consultaId, para não expor todos os documentos.
func (s *Documentos) Listar(ctx context.Context, f domain.FiltroDocumentos) ([]domain.Documento, error) {
	if f.PacienteID == nil && f.ConsultaID == nil {
		return nil, &domain.ErroValidacao{Mensagem: "informe pacienteId ou consultaId"}
	}
	docs, err := s.repo.Listar(ctx, f)
	if docs == nil && err == nil {
		docs = []domain.Documento{}
	}
	return docs, err
}

func (s *Documentos) Baixar(ctx context.Context, id uuid.UUID) (domain.Documento, io.ReadCloser, error) {
	doc, err := s.repo.BuscarPorID(ctx, id)
	if err != nil {
		return domain.Documento{}, nil, err
	}
	conteudo, err := s.armazenamento.Abrir(ctx, doc.URI)
	if err != nil {
		return domain.Documento{}, nil, err
	}
	return doc, conteudo, nil
}

func (s *Documentos) validar(n domain.NovoDocumento) error {
	if _, err := domain.ParseTipoDocumento(string(n.Tipo)); err != nil {
		return err
	}
	if n.PacienteID == uuid.Nil {
		return &domain.ErroValidacao{Mensagem: "pacienteId é obrigatório"}
	}
	if n.AutorID == uuid.Nil {
		return &domain.ErroValidacao{Mensagem: "autorId é obrigatório"}
	}
	if n.ConsultaID != nil && *n.ConsultaID == uuid.Nil {
		return &domain.ErroValidacao{Mensagem: "consultaId inválido"}
	}
	if n.NomeOriginal == "" {
		return &domain.ErroValidacao{Mensagem: "o arquivo precisa ter um nome"}
	}
	if n.TamanhoBytes <= 0 {
		return &domain.ErroValidacao{Mensagem: "o arquivo está vazio"}
	}
	if n.TamanhoBytes > s.tamanhoMax {
		return domain.ErrArquivoMuitoGrande
	}
	if !domain.FormatosPermitidos[n.ContentType] {
		return domain.ErrFormatoNaoPermitido
	}
	return nil
}

// limparNome descarta diretórios vindos do cliente e limita o tamanho.
func limparNome(nome string) string {
	nome = strings.ToValidUTF8(nome, "")
	nome = strings.ReplaceAll(nome, `\`, "/")
	nome = strings.TrimSpace(path.Base(path.Clean("/" + nome)))
	if nome == "/" {
		return ""
	}
	if len(nome) > tamanhoMaxNome {
		// Corta em bytes e descarta um eventual caractere UTF-8 partido ao meio.
		nome = strings.ToValidUTF8(nome[:tamanhoMaxNome], "")
	}
	return nome
}
