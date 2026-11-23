// Package memoria implementa o ProntuarioRepository em memória, no mesmo
// espírito do ConsultaRepositoryMemoria do diagrama. Usado nos testes.
package memoria

import (
	"context"
	"slices"
	"sync"

	"github.com/google/uuid"

	"prontuario-service/internal/domain"
)

type ProntuarioRepository struct {
	mu          sync.Mutex
	prontuarios map[uuid.UUID]domain.Prontuario // chave: pacienteID
	documentos  map[uuid.UUID][]domain.DocumentoAnexado
	laudos      map[uuid.UUID][]domain.Laudo
}

func NovoProntuarioRepository() *ProntuarioRepository {
	return &ProntuarioRepository{
		prontuarios: make(map[uuid.UUID]domain.Prontuario),
		documentos:  make(map[uuid.UUID][]domain.DocumentoAnexado),
		laudos:      make(map[uuid.UUID][]domain.Laudo),
	}
}

func (r *ProntuarioRepository) ObterOuCriar(_ context.Context, novo domain.Prontuario) (domain.Prontuario, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.prontuarios[novo.PacienteID]; ok {
		return p, nil
	}
	r.prontuarios[novo.PacienteID] = novo
	return novo, nil
}

func (r *ProntuarioRepository) AnexarDocumento(_ context.Context, prontuarioID uuid.UUID, d domain.DocumentoAnexado) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if slices.ContainsFunc(r.documentos[prontuarioID], func(x domain.DocumentoAnexado) bool { return x.DocumentoID == d.DocumentoID }) {
		return false, nil
	}
	r.documentos[prontuarioID] = append(r.documentos[prontuarioID], d)
	return true, nil
}

func (r *ProntuarioRepository) ListarDocumentos(_ context.Context, prontuarioID uuid.UUID) ([]domain.DocumentoAnexado, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.documentos[prontuarioID]), nil
}

func (r *ProntuarioRepository) SalvarLaudo(_ context.Context, l domain.Laudo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.laudos[l.ProntuarioID] = append(r.laudos[l.ProntuarioID], l)
	return nil
}

func (r *ProntuarioRepository) ListarLaudos(_ context.Context, prontuarioID uuid.UUID) ([]domain.Laudo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.laudos[prontuarioID]), nil
}
