package prontuario

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"documentos-service/internal/domain"
)

func TestAnexarDocumento(t *testing.T) {
	doc := domain.Documento{
		ID: uuid.New(), Tipo: domain.TipoLaudo, PacienteID: uuid.New(), AutorID: uuid.New(),
		NomeOriginal: "laudo.pdf", ContentType: "application/pdf", TamanhoBytes: 10,
		EnviadoEm: time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC),
	}

	status := http.StatusCreated
	var recebido anexoRequisicao
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		quer := "/pacientes/" + doc.PacienteID.String() + "/prontuario/documentos"
		if r.Method != http.MethodPost || r.URL.Path != quer {
			t.Errorf("requisição %s %s, quer POST %s", r.Method, r.URL.Path, quer)
		}
		json.NewDecoder(r.Body).Decode(&recebido)
		w.WriteHeader(status)
		w.Write([]byte(`{"erro":"detalhe"}`))
	}))
	defer srv.Close()
	c := NovoCliente(srv.URL)

	if err := c.AnexarDocumento(context.Background(), doc); err != nil {
		t.Fatalf("201: %v", err)
	}
	if recebido.DocumentoID != doc.ID || recebido.Tipo != "LAUDO" || !recebido.EnviadoEm.Equal(doc.EnviadoEm) {
		t.Errorf("corpo enviado inesperado: %+v", recebido)
	}

	status = http.StatusOK // já anexado: também é sucesso
	if err := c.AnexarDocumento(context.Background(), doc); err != nil {
		t.Errorf("200: %v", err)
	}

	status = http.StatusBadRequest
	if err := c.AnexarDocumento(context.Background(), doc); err == nil {
		t.Error("400 deveria ser erro")
	}

	srv.Close()
	if err := c.AnexarDocumento(context.Background(), doc); err == nil {
		t.Error("serviço fora do ar deveria ser erro")
	}
}
