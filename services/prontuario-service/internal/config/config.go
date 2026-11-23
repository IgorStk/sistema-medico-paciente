// Package config lê a configuração do serviço a partir de variáveis de ambiente.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Porta                string
	DatabaseURL          string
	CORSOrigens          []string
	DocumentosServiceURL string
}

// Load lê as variáveis de ambiente. Os valores padrão servem para rodar
// localmente com o docker-compose da raiz do repositório.
func Load() (Config, error) {
	cfg := Config{
		Porta:                env("PORT", "8082"),
		DatabaseURL:          env("DATABASE_URL", "postgres://prontuario:prontuario@localhost:5434/prontuario_db?sslmode=disable"),
		CORSOrigens:          separarLista(env("CORS_ORIGINS", "http://localhost:5173")),
		DocumentosServiceURL: strings.TrimRight(env("DOCUMENTOS_SERVICE_URL", "http://localhost:8081"), "/"),
	}

	u, err := url.Parse(cfg.DocumentosServiceURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Config{}, fmt.Errorf("DOCUMENTOS_SERVICE_URL inválida: %q", cfg.DocumentosServiceURL)
	}
	return cfg, nil
}

func env(chave, padrao string) string {
	if v, ok := os.LookupEnv(chave); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return padrao
}

func separarLista(s string) []string {
	var itens []string
	for _, parte := range strings.Split(s, ",") {
		if p := strings.TrimSpace(parte); p != "" {
			itens = append(itens, p)
		}
	}
	return itens
}
