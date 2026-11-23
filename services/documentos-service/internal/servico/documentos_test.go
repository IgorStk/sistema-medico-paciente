package servico

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"

	"documentos-service/internal/adapters/memoria"
	"documentos-service/internal/domain"
)

const limite = 1 << 20 // 1 MB

func novoServicoTeste() (*Documentos, *memoria.DocumentoRepository, *memoria.Armazenamento) {
	repo := memoria.NovoDocumentoRepository()
	arm := memoria.NovoArmazenamento()
	return NovoDocumentos(repo, arm, limite), repo, arm
}

func novoValido() domain.NovoDocumento {
	return domain.NovoDocumento{
		Tipo:         domain.TipoExamePrevio,
		PacienteID:   uuid.New(),
		AutorID:      uuid.New(),
		NomeOriginal: "hemograma.pdf",
		ContentType:  "application/pdf",
		TamanhoBytes: 9,
	}
}

func TestEnviarEBaixar(t *testing.T) {
	s, _, arm := novoServicoTeste()
	ctx := context.Background()

	doc, err := s.Enviar(ctx, novoValido(), strings.NewReader("%PDF-1.4\n"))
	if err != nil {
		t.Fatal(err)
	}
	if doc.ID == uuid.Nil || doc.EnviadoEm.IsZero() || !arm.Existe(doc.URI) {
		t.Fatalf("documento incompleto ou arquivo não salvo: %+v", doc)
	}

	baixado, conteudo, err := s.Baixar(ctx, doc.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer conteudo.Close()
	dados, _ := io.ReadAll(conteudo)
	if baixado.ID != doc.ID || string(dados) != "%PDF-1.4\n" {
		t.Errorf("download diferente do upload: %+v %q", baixado, dados)
	}
}

func TestEnviarValidacoes(t *testing.T) {
	casos := []struct {
		nome    string
		mudar   func(*domain.NovoDocumento)
		querIs  error
		quer400 bool
	}{
		{"tipo inválido", func(n *domain.NovoDocumento) { n.Tipo = "RAIO_X" }, domain.ErrTipoInvalido, false},
		{"sem paciente", func(n *domain.NovoDocumento) { n.PacienteID = uuid.Nil }, nil, true},
		{"sem autor", func(n *domain.NovoDocumento) { n.AutorID = uuid.Nil }, nil, true},
		{"vazio", func(n *domain.NovoDocumento) { n.TamanhoBytes = 0 }, nil, true},
		{"sem nome", func(n *domain.NovoDocumento) { n.NomeOriginal = "  " }, nil, true},
		{"grande demais", func(n *domain.NovoDocumento) { n.TamanhoBytes = limite + 1 }, domain.ErrArquivoMuitoGrande, false},
		{"formato proibido", func(n *domain.NovoDocumento) { n.ContentType = "application/zip" }, domain.ErrFormatoNaoPermitido, false},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			s, _, _ := novoServicoTeste()
			n := novoValido()
			c.mudar(&n)
			_, err := s.Enviar(context.Background(), n, strings.NewReader("x"))

			var ev *domain.ErroValidacao
			switch {
			case c.querIs != nil && !errors.Is(err, c.querIs):
				t.Errorf("erro = %v, quer %v", err, c.querIs)
			case c.quer400 && !errors.As(err, &ev):
				t.Errorf("erro = %v, quer ErroValidacao", err)
			}
		})
	}
}

func TestEnviarRemoveArquivoSeBancoFalhar(t *testing.T) {
	s, repo, arm := novoServicoTeste()
	repo.FalharSalvar = errors.New("banco fora do ar")

	_, err := s.Enviar(context.Background(), novoValido(), strings.NewReader("%PDF-1.4\n"))
	if err == nil {
		t.Fatal("esperava erro")
	}
	// Nenhum objeto deve ter ficado no armazenamento.
	if arm.Total() != 0 {
		t.Errorf("arquivo órfão ficou no armazenamento (%d objetos)", arm.Total())
	}
}

func TestListar(t *testing.T) {
	s, _, _ := novoServicoTeste()
	ctx := context.Background()

	if _, err := s.Listar(ctx, domain.FiltroDocumentos{}); err == nil {
		t.Error("listar sem filtro deveria falhar")
	}

	n := novoValido()
	consulta := uuid.New()
	n.ConsultaID = &consulta
	if _, err := s.Enviar(ctx, n, strings.NewReader("%PDF-1.4\n")); err != nil {
		t.Fatal(err)
	}
	outro := novoValido() // outro paciente, sem consulta
	if _, err := s.Enviar(ctx, outro, strings.NewReader("%PDF-1.4\n")); err != nil {
		t.Fatal(err)
	}

	porConsulta, err := s.Listar(ctx, domain.FiltroDocumentos{ConsultaID: &consulta})
	if err != nil || len(porConsulta) != 1 {
		t.Errorf("por consulta: %d docs, erro %v; quer 1", len(porConsulta), err)
	}

	ninguem := uuid.New()
	vazio, err := s.Listar(ctx, domain.FiltroDocumentos{PacienteID: &ninguem})
	if err != nil || vazio == nil || len(vazio) != 0 {
		t.Errorf("sem resultados deveria ser lista vazia não-nil: %#v, %v", vazio, err)
	}
}

func TestLimparNome(t *testing.T) {
	casos := map[string]string{
		"exame.pdf":              "exame.pdf",
		`C:\Users\ana\exame.pdf`: "exame.pdf",
		"../../etc/passwd":       "passwd",
		"  laudo final.pdf  ":    "laudo final.pdf",
		"..":                     "",
		"":                       "",
		strings.Repeat("é", 200): strings.Repeat("é", 127),
	}
	for entrada, quer := range casos {
		if obtido := limparNome(entrada); obtido != quer {
			t.Errorf("limparNome(%q) = %q, quer %q", entrada, obtido, quer)
		}
	}
}
