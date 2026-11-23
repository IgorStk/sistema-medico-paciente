package servico

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"prontuario-service/internal/adapters/memoria"
	"prontuario-service/internal/domain"
)

func docValido(enviadoEm time.Time) domain.DocumentoAnexado {
	return domain.DocumentoAnexado{
		DocumentoID:  uuid.New(),
		Tipo:         "EXAME_PREVIO",
		NomeOriginal: "hemograma.pdf",
		ContentType:  "application/pdf",
		TamanhoBytes: 1234,
		AutorID:      uuid.New(),
		EnviadoEm:    enviadoEm,
	}
}

func TestObterOuCriarDevolveSempreOMesmo(t *testing.T) {
	s := NovoProntuario(memoria.NovoProntuarioRepository())
	paciente := uuid.New()

	p1, err := s.ObterOuCriar(context.Background(), paciente)
	if err != nil {
		t.Fatal(err)
	}
	p2, _ := s.ObterOuCriar(context.Background(), paciente)
	if p1.ID != p2.ID || p1.PacienteID != paciente {
		t.Errorf("prontuários diferentes para o mesmo paciente: %v %v", p1, p2)
	}

	if _, err := s.ObterOuCriar(context.Background(), uuid.Nil); err == nil {
		t.Error("pacienteId nulo deveria falhar")
	}
}

func TestAnexarEhIdempotente(t *testing.T) {
	s := NovoProntuario(memoria.NovoProntuarioRepository())
	ctx, paciente := context.Background(), uuid.New()
	doc := docValido(time.Now())

	_, novo, err := s.AnexarDocumento(ctx, paciente, doc)
	if err != nil || !novo {
		t.Fatalf("primeiro anexo: novo=%v err=%v", novo, err)
	}
	_, novo, err = s.AnexarDocumento(ctx, paciente, doc)
	if err != nil || novo {
		t.Fatalf("segundo anexo do mesmo documento: novo=%v err=%v; quer novo=false", novo, err)
	}

	hist, _ := s.Historico(ctx, paciente)
	if len(hist) != 1 {
		t.Errorf("histórico com %d itens, quer 1", len(hist))
	}
}

func TestAnexarValida(t *testing.T) {
	s := NovoProntuario(memoria.NovoProntuarioRepository())
	doc := docValido(time.Now())
	doc.Tipo = "RAIO_X"
	if _, _, err := s.AnexarDocumento(context.Background(), uuid.New(), doc); !errors.Is(err, domain.ErrTipoDocumentoInvalido) {
		t.Errorf("erro = %v, quer ErrTipoDocumentoInvalido", err)
	}

	doc = docValido(time.Time{})
	var ev *domain.ErroValidacao
	if _, _, err := s.AnexarDocumento(context.Background(), uuid.New(), doc); !errors.As(err, &ev) {
		t.Errorf("sem enviadoEm: erro = %v, quer ErroValidacao", err)
	}
}

func TestHistoricoJuntaDocumentosELaudosEmOrdem(t *testing.T) {
	s := NovoProntuario(memoria.NovoProntuarioRepository())
	ctx, paciente := context.Background(), uuid.New()
	base := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)

	// Laudo emitido "agora" (depois dos documentos)
	s.agora = func() time.Time { return base.Add(48 * time.Hour) }
	if _, err := s.EmitirLaudo(ctx, paciente, domain.NovoLaudo{MedicoID: uuid.New(), Titulo: "Laudo", Conteudo: "Normal."}); err != nil {
		t.Fatal(err)
	}
	antigo, recente := docValido(base), docValido(base.Add(24*time.Hour))
	s.AnexarDocumento(ctx, paciente, antigo)
	s.AnexarDocumento(ctx, paciente, recente)

	hist, err := s.Historico(ctx, paciente)
	if err != nil {
		t.Fatal(err)
	}
	if len(hist) != 3 || hist[0].Tipo != "LAUDO" ||
		hist[1].Documento.DocumentoID != recente.DocumentoID || hist[2].Documento.DocumentoID != antigo.DocumentoID {
		t.Errorf("ordem inesperada: %+v", hist)
	}
}

func TestEmitirLaudoValida(t *testing.T) {
	s := NovoProntuario(memoria.NovoProntuarioRepository())
	medico := uuid.New()
	casos := map[string]domain.NovoLaudo{
		"sem médico":    {Titulo: "T", Conteudo: "C"},
		"sem título":    {MedicoID: medico, Titulo: "  ", Conteudo: "C"},
		"sem conteúdo":  {MedicoID: medico, Titulo: "T"},
		"título enorme": {MedicoID: medico, Titulo: strings.Repeat("a", domain.TamanhoMaxTitulo+1), Conteudo: "C"},
		"consulta nula": {MedicoID: medico, Titulo: "T", Conteudo: "C", ConsultaID: &uuid.Nil},
	}
	for nome, n := range casos {
		var ev *domain.ErroValidacao
		if _, err := s.EmitirLaudo(context.Background(), uuid.New(), n); !errors.As(err, &ev) {
			t.Errorf("%s: erro = %v, quer ErroValidacao", nome, err)
		}
	}
}
