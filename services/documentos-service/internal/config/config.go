// Package config lê a configuração do serviço a partir de variáveis de ambiente.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Porta              string
	DatabaseURL        string
	CORSOrigens        []string
	TamanhoMaxUploadMB int64
}

// Load lê as variáveis de ambiente. Os valores padrão servem para rodar
// localmente com o docker-compose da raiz do repositório.
func Load() (Config, error) {
	cfg := Config{
		Porta:       env("PORT", "8081"),
		DatabaseURL: env("DATABASE_URL", "postgres://documentos:documentos@localhost:5433/documentos_db?sslmode=disable"),
		CORSOrigens: separarLista(env("CORS_ORIGINS", "http://localhost:5173")),
	}

	tamanho, err := strconv.ParseInt(env("MAX_UPLOAD_MB", "20"), 10, 64)
	if err != nil || tamanho <= 0 {
		return Config{}, fmt.Errorf("MAX_UPLOAD_MB deve ser um inteiro positivo")
	}
	cfg.TamanhoMaxUploadMB = tamanho

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
