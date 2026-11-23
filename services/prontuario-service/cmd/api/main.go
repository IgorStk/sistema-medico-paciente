// Comando api sobe o prontuario-service: prontuário do paciente, anexos,
// histórico e emissão de laudos. Consome o documentos-service via HTTP.
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

	"prontuario-service/internal/adapters/documentos"
	"prontuario-service/internal/adapters/httpapi"
	"prontuario-service/internal/adapters/postgres"
	"prontuario-service/internal/config"
)

const nomeServico = "prontuario-service"

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

	pool, err := postgres.Conectar(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := postgres.Migrar(ctx, pool); err != nil {
		return err
	}
	log.Info("banco conectado e migrações aplicadas")

	clienteDocumentos := documentos.NovoCliente(cfg.DocumentosServiceURL)

	servidor := &http.Server{
		Addr:              ":" + cfg.Porta,
		Handler:           httpapi.NovoRouter(httpapi.NovoHandler(nomeServico, pool, clienteDocumentos), cfg.CORSOrigens, log),
		ReadHeaderTimeout: 5 * time.Second,
	}

	erros := make(chan error, 1)
	go func() {
		log.Info("servidor HTTP ouvindo", "porta", cfg.Porta, "documentos_service", cfg.DocumentosServiceURL)
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
