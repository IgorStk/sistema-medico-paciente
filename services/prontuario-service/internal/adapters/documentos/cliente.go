// Package documentos é o cliente HTTP do documentos-service. É o adaptador que
// implementa a porta ports.DocumentosGateway (BuscarDocumento e
// EnviarDocumento entram na fase 5).
package documentos

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Cliente struct {
	baseURL string
	http    *http.Client
}

func NovoCliente(baseURL string) *Cliente {
	return &Cliente{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Ping consulta o /health do documentos-service.
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
		return fmt.Errorf("documentos-service respondeu %d", resp.StatusCode)
	}
	return nil
}
