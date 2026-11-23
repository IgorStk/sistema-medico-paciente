// Comando api sobe o documentos-service: upload, consulta e download de
// documentos clínicos (exames prévios, laudos, receitas e atestados).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"documentos-service/internal/adapters/httpapi"
	"documentos-service/internal/adapters/postgres"
	"documentos-service/internal/adapters/prontuario"
	"documentos-service/internal/adapters/s3"
	"documentos-service/internal/config"
	"documentos-service/internal/servico"
)

const nomeServico = "documentos-service"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("servico", nomeServico)
	if err := executar(log); err != nil {
		log.Error("serviço encerrado com erro", "erro", err)
		os.Exit(1)
	}
}

func executar(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	// Banco de metadados
	pool, err := postgres.Conectar(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := postgres.Migrar(ctx, pool); err != nil {
		return err
	}
	log.Info("banco conectado e migrações aplicadas")

	// Armazenamento dos arquivos (S3)
	armazenamento := s3.Novo(s3.Config(cfg.S3))
	if err := armazenamento.GarantirBucket(ctx); err != nil {
		return err
	}
	log.Info("armazenamento S3 pronto", "endpoint", cfg.S3.Endpoint, "bucket", cfg.S3.Bucket)

	// Casos de uso e API
	tamanhoMax := cfg.TamanhoMaxUploadMB << 20
	clienteProntuario := prontuario.NovoCliente(cfg.ProntuarioServiceURL)
	documentos := servico.NovoDocumentos(postgres.NovoDocumentoRepository(pool), armazenamento, clienteProntuario, tamanhoMax)
	handler := httpapi.NovoHandler(httpapi.Dependencias{
		Servico:               nomeServico,
		Log:                   log,
		Banco:                 pool,
		Armazenamento:         armazenamento,
		Prontuario:            clienteProntuario,
		Upload:                documentos,
		Download:              documentos,
		TamanhoMaxUploadBytes: tamanhoMax,
	})

	servidor := &http.Server{
		Addr:              ":" + cfg.Porta,
		Handler:           httpapi.NovoRouter(handler, cfg.CORSOrigens),
		ReadHeaderTimeout: 5 * time.Second,
	}

	erros := make(chan error, 1)
	go func() {
		log.Info("servidor HTTP ouvindo", "porta", cfg.Porta, "prontuario_service", cfg.ProntuarioServiceURL)
		if err := servidor.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			erros <- err
		}
	}()

	select {
	case err := <-erros:
		return err
	case <-ctx.Done():
	}

	log.Info("desligando o servidor")
	ctxDesligar, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	return servidor.Shutdown(ctxDesligar)
}
