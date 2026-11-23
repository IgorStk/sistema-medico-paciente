// Package config lê a configuração do serviço a partir de variáveis de ambiente.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Porta              string
	DatabaseURL        string
	CORSOrigens        []string
	TamanhoMaxUploadMB int64
	S3                 S3
}

// S3 configura o armazenamento dos arquivos. Os padrões apontam para o S3
// local do docker-compose (SeaweedFS) e servem só para desenvolvimento.
type S3 struct {
	Endpoint  string // URL completa, ex.: https://<projeto>.storage.supabase.co/storage/v1/s3
	AccessKey string
	SecretKey string
	Bucket    string
	Regiao    string
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

	cfg.S3 = S3{
		Endpoint:  strings.TrimRight(env("S3_ENDPOINT", "http://localhost:8333"), "/"),
		AccessKey: env("S3_ACCESS_KEY", "documentos"),
		SecretKey: env("S3_SECRET_KEY", "documentos-segredo-dev"),
		Bucket:    env("S3_BUCKET", "documentos"),
		Regiao:    env("S3_REGION", "us-east-1"),
	}
	u, err := url.Parse(cfg.S3.Endpoint)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Config{}, fmt.Errorf("S3_ENDPOINT deve ser uma URL completa (http:// ou https://), veio %q", cfg.S3.Endpoint)
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
