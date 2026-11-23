// Package memoria tem implementações em memória das portas de saída, no
// mesmo espírito do ConsultaRepositoryMemoria do diagrama. Servem para os
// testes rodarem sem Postgres nem S3.
package memoria

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"slices"
	"sync"

	"github.com/google/uuid"

	"documentos-service/internal/domain"
)

// DocumentoRepository implementa ports.DocumentoRepository em memória.
type DocumentoRepository struct {
	mu   sync.Mutex
	docs map[uuid.UUID]domain.Documento
	// FalharSalvar, se definido, é devolvido por Salvar (para testar falhas).
	FalharSalvar error
}

func NovoDocumentoRepository() *DocumentoRepository {
	return &DocumentoRepository{docs: make(map[uuid.UUID]domain.Documento)}
}

func (r *DocumentoRepository) Salvar(_ context.Context, d domain.Documento) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.FalharSalvar != nil {
		return r.FalharSalvar
	}
	r.docs[d.ID] = d
	return nil
}

func (r *DocumentoRepository) BuscarPorID(_ context.Context, id uuid.UUID) (domain.Documento, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.docs[id]
	if !ok {
		return domain.Documento{}, domain.ErrNaoEncontrado
	}
	return d, nil
}

func (r *DocumentoRepository) Listar(_ context.Context, f domain.FiltroDocumentos) ([]domain.Documento, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var docs []domain.Documento
	for _, d := range r.docs {
		if f.PacienteID != nil && d.PacienteID != *f.PacienteID {
			continue
		}
		if f.ConsultaID != nil && (d.ConsultaID == nil || *d.ConsultaID != *f.ConsultaID) {
			continue
		}
		if f.Tipo != nil && d.Tipo != *f.Tipo {
			continue
		}
		docs = append(docs, d)
	}
	slices.SortFunc(docs, func(a, b domain.Documento) int { return b.EnviadoEm.Compare(a.EnviadoEm) })
	return docs, nil
}

func (r *DocumentoRepository) Remover(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.docs, id)
	return nil
}

// Prontuario implementa ports.ProntuarioGateway em memória: guarda os
// documentos anexados por paciente.
type Prontuario struct {
	mu     sync.Mutex
	anexos map[uuid.UUID][]uuid.UUID
	Falhar error // se definido, AnexarDocumento devolve este erro
}

func NovoProntuario() *Prontuario {
	return &Prontuario{anexos: make(map[uuid.UUID][]uuid.UUID)}
}

func (p *Prontuario) AnexarDocumento(_ context.Context, d domain.Documento) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Falhar != nil {
		return p.Falhar
	}
	p.anexos[d.PacienteID] = append(p.anexos[d.PacienteID], d.ID)
	return nil
}

// Anexos devolve os documentos anexados ao prontuário do paciente.
func (p *Prontuario) Anexos(pacienteID uuid.UUID) []uuid.UUID {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.anexos[pacienteID])
}

// Armazenamento implementa ports.ArmazenamentoArquivos em memória.
type Armazenamento struct {
	mu       sync.Mutex
	arquivos map[string][]byte
}

func NovoArmazenamento() *Armazenamento {
	return &Armazenamento{arquivos: make(map[string][]byte)}
}

func (a *Armazenamento) Salvar(_ context.Context, chave string, conteudo io.Reader, _ int64, _ string) error {
	dados, err := io.ReadAll(conteudo)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.arquivos[chave] = dados
	return nil
}

func (a *Armazenamento) Abrir(_ context.Context, chave string) (io.ReadCloser, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	dados, ok := a.arquivos[chave]
	if !ok {
		return nil, fmt.Errorf("objeto %q não existe", chave)
	}
	return io.NopCloser(bytes.NewReader(dados)), nil
}

func (a *Armazenamento) Remover(_ context.Context, chave string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.arquivos, chave)
	return nil
}

// Existe informa se há um objeto com a chave (uso nos testes).
func (a *Armazenamento) Existe(chave string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, ok := a.arquivos[chave]
	return ok
}

// Total devolve quantos objetos estão guardados (uso nos testes).
func (a *Armazenamento) Total() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.arquivos)
}
