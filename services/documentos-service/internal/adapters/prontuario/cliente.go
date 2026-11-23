// Package prontuario é o cliente HTTP do prontuario-service. Implementa a
// porta ports.ProntuarioGateway (seta "anexar" do diagrama de microsserviços).
package prontuario

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"documentos-service/internal/domain"
)

type Cliente struct {
	baseURL string
	http    *http.Client
}

func NovoCliente(baseURL string) *Cliente {
	return &Cliente{baseURL: baseURL, http: &http.Client{Timeout: 10 * time.Second}}
}

// anexoRequisicao é o contrato de POST /pacientes/{id}/prontuario/documentos.
type anexoRequisicao struct {
	DocumentoID  uuid.UUID  `json:"documentoId"`
	Tipo         string     `json:"tipo"`
	NomeOriginal string     `json:"nomeOriginal"`
	ContentType  string     `json:"contentType"`
	TamanhoBytes int64      `json:"tamanhoBytes"`
	ConsultaID   *uuid.UUID `json:"consultaId,omitempty"`
	AutorID      uuid.UUID  `json:"autorId"`
	EnviadoEm    time.Time  `json:"enviadoEm"`
}

// AnexarDocumento pede ao prontuário que anexe o documento. 200 e 201 contam
// como sucesso (o prontuário não duplica anexos repetidos).
func (c *Cliente) AnexarDocumento(ctx context.Context, d domain.Documento) error {
	corpo, err := json.Marshal(anexoRequisicao{
		DocumentoID:  d.ID,
		Tipo:         string(d.Tipo),
		NomeOriginal: d.NomeOriginal,
		ContentType:  d.ContentType,
		TamanhoBytes: d.TamanhoBytes,
		ConsultaID:   d.ConsultaID,
		AutorID:      d.AutorID,
		EnviadoEm:    d.EnviadoEm,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/pacientes/%s/prontuario/documentos", c.baseURL, d.PacienteID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(corpo))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("chamando prontuario-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		detalhe, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("prontuario-service respondeu %d: %s", resp.StatusCode, bytes.TrimSpace(detalhe))
	}
	return nil
}

// Ping consulta o /health do prontuario-service.
func (c *Cliente) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("prontuario-service respondeu %d", resp.StatusCode)
	}
	return nil
}
