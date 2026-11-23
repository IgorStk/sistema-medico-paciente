package httpapi

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"documentos-service/internal/domain"
)

var pdfMinimo = []byte("%PDF-1.4\n1 0 obj << >> endobj\ntrailer << >>\n%%EOF\n")

// requisicaoUpload monta um POST /documentos multipart. Campos com valor ""
// são omitidos; arquivo nil omite o campo arquivo.
func requisicaoUpload(t *testing.T, campos map[string]string, nomeArquivo string, arquivo []byte) *http.Request {
	t.Helper()
	var corpo bytes.Buffer
	mw := multipart.NewWriter(&corpo)
	for k, v := range campos {
		if v != "" {
			_ = mw.WriteField(k, v)
		}
	}
	if arquivo != nil {
		parte, err := mw.CreateFormFile("arquivo", nomeArquivo)
		if err != nil {
			t.Fatal(err)
		}
		parte.Write(arquivo)
	}
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/documentos", &corpo)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func camposValidos() map[string]string {
	return map[string]string{
		"tipo":       "EXAME_PREVIO",
		"pacienteId": uuid.NewString(),
		"autorId":    uuid.NewString(),
		"consultaId": uuid.NewString(),
	}
}

func TestFluxoCompletoUploadListagemDownload(t *testing.T) {
	r := novoRouterTeste(nil, nil)
	campos := camposValidos()

	// Upload
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, requisicaoUpload(t, campos, "hemograma ação.pdf", pdfMinimo))
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload: status %d, corpo %s", rec.Code, rec.Body)
	}
	var doc domain.Documento
	if err := json.NewDecoder(rec.Body).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	if doc.ContentType != "application/pdf" || doc.NomeOriginal != "hemograma ação.pdf" ||
		doc.TamanhoBytes != int64(len(pdfMinimo)) || doc.PacienteID.String() != campos["pacienteId"] {
		t.Errorf("documento inesperado: %+v", doc)
	}
	if loc := rec.Header().Get("Location"); loc != "/documentos/"+doc.ID.String() {
		t.Errorf("Location = %q", loc)
	}

	// Metadados
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/documentos/"+doc.ID.String(), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("buscar: status %d", rec.Code)
	}

	// Listagem por consulta
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/documentos?consultaId="+campos["consultaId"], nil))
	var lista []domain.Documento
	if err := json.NewDecoder(rec.Body).Decode(&lista); err != nil || len(lista) != 1 {
		t.Fatalf("listar: status %d, %d itens, erro %v", rec.Code, len(lista), err)
	}

	// Download
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/documentos/"+doc.ID.String()+"/arquivo", nil))
	if rec.Code != http.StatusOK || !bytes.Equal(rec.Body.Bytes(), pdfMinimo) {
		t.Fatalf("download: status %d, %d bytes", rec.Code, rec.Body.Len())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Content-Type = %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); cd != `attachment; filename*=utf-8''hemograma%20a%C3%A7%C3%A3o.pdf` {
		t.Errorf("Content-Disposition = %q", cd)
	}
}

func TestUploadErros(t *testing.T) {
	r := novoRouterTeste(nil, nil)

	casos := []struct {
		nome    string
		mudar   func(map[string]string)
		arquivo []byte
		quer    int
	}{
		{"tipo inválido", func(c map[string]string) { c["tipo"] = "RAIO_X" }, pdfMinimo, http.StatusBadRequest},
		{"sem paciente", func(c map[string]string) { c["pacienteId"] = "" }, pdfMinimo, http.StatusBadRequest},
		{"paciente não é uuid", func(c map[string]string) { c["pacienteId"] = "123" }, pdfMinimo, http.StatusBadRequest},
		{"sem arquivo", func(map[string]string) {}, nil, http.StatusBadRequest},
		{"arquivo vazio", func(map[string]string) {}, []byte{}, http.StatusBadRequest},
		{"formato proibido", func(map[string]string) {}, []byte("PK\x03\x04 isto é um zip"), http.StatusUnsupportedMediaType},
		{"grande demais", func(map[string]string) {}, append([]byte("%PDF-1.4\n"), make([]byte, limiteTeste)...), http.StatusRequestEntityTooLarge},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			campos := camposValidos()
			c.mudar(campos)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, requisicaoUpload(t, campos, "arquivo.pdf", c.arquivo))
			if rec.Code != c.quer {
				t.Errorf("status = %d, quer %d (corpo %s)", rec.Code, c.quer, rec.Body)
			}
		})
	}
}

func TestUploadSemMultipart(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/documentos", bytes.NewBufferString(`{"tipo":"LAUDO"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	novoRouterTeste(nil, nil).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, quer 400", rec.Code)
	}
}

func TestConsultasComErro(t *testing.T) {
	r := novoRouterTeste(nil, nil)
	casos := []struct {
		caminho string
		quer    int
	}{
		{"/documentos/" + uuid.NewString(), http.StatusNotFound},
		{"/documentos/" + uuid.NewString() + "/arquivo", http.StatusNotFound},
		{"/documentos/nao-e-uuid", http.StatusBadRequest},
		{"/documentos", http.StatusBadRequest}, // listar sem filtro
		{"/documentos?pacienteId=xyz", http.StatusBadRequest},
		{"/documentos?pacienteId=" + uuid.NewString() + "&tipo=FOTO", http.StatusBadRequest},
	}
	for _, c := range casos {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, c.caminho, nil))
		if rec.Code != c.quer {
			t.Errorf("GET %s = %d, quer %d", c.caminho, rec.Code, c.quer)
		}
	}

	// Lista vazia deve ser [] e não null, para o frontend não quebrar.
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/documentos?pacienteId="+uuid.NewString(), nil))
	if corpo := bytes.TrimSpace(rec.Body.Bytes()); string(corpo) != "[]" {
		t.Errorf("lista vazia = %s, quer []", corpo)
	}
}
