// Package s3 contém o ArmazenamentoS3 do diagrama: guarda o conteúdo dos
// documentos em um bucket compatível com S3. Usa o SDK oficial da AWS, que
// funciona com o Amazon S3, o Supabase Storage (produção) e o SeaweedFS
// (docker-compose local).
package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type Config struct {
	// Endpoint é a URL completa da API S3, por exemplo:
	//   Supabase:  https://<projeto>.storage.supabase.co/storage/v1/s3
	//   local:     http://localhost:8333
	//   Amazon S3: vazio (o SDK usa o endpoint da região)
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Regiao    string
}

// ArmazenamentoS3 implementa ports.ArmazenamentoArquivos.
type ArmazenamentoS3 struct {
	cliente *awss3.Client
	bucket  string
	regiao  string
}

func Novo(cfg Config) *ArmazenamentoS3 {
	opcoes := awss3.Options{
		Region:      cfg.Regiao,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		// Endereço no formato <endpoint>/<bucket>/<chave>, aceito por todos os provedores.
		UsePathStyle: true,
		// Checksums só quando a operação exige: provedores compatíveis com S3
		// nem sempre aceitam os checksums extras que o SDK manda por padrão.
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	}
	if cfg.Endpoint != "" {
		opcoes.BaseEndpoint = aws.String(cfg.Endpoint)
	}
	return &ArmazenamentoS3{cliente: awss3.New(opcoes), bucket: cfg.Bucket, regiao: cfg.Regiao}
}

// GarantirBucket cria o bucket se ainda não existir. Tenta por alguns
// segundos, porque no docker-compose o S3 local pode demorar a subir.
func (a *ArmazenamentoS3) GarantirBucket(ctx context.Context) error {
	var err error
	for tentativa := 1; tentativa <= 30; tentativa++ {
		if err = a.criarSeFaltar(ctx); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return fmt.Errorf("armazenamento S3 indisponível: %w", err)
}

func (a *ArmazenamentoS3) criarSeFaltar(ctx context.Context) error {
	err := a.Ping(ctx)
	if err == nil || !codigoErro(err, "NotFound", "NoSuchBucket") {
		return err
	}

	entrada := &awss3.CreateBucketInput{Bucket: aws.String(a.bucket)}
	if a.regiao != "" && a.regiao != "us-east-1" {
		entrada.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(a.regiao),
		}
	}
	_, err = a.cliente.CreateBucket(ctx, entrada)
	if codigoErro(err, "BucketAlreadyOwnedByYou", "BucketAlreadyExists") {
		return nil
	}
	return err
}

// Ping verifica se o bucket existe e se as credenciais funcionam (/health).
func (a *ArmazenamentoS3) Ping(ctx context.Context) error {
	_, err := a.cliente.HeadBucket(ctx, &awss3.HeadBucketInput{Bucket: aws.String(a.bucket)})
	return err
}

func (a *ArmazenamentoS3) Salvar(ctx context.Context, chave string, conteudo io.Reader, tamanho int64, contentType string) error {
	_, err := a.cliente.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:        aws.String(a.bucket),
		Key:           aws.String(chave),
		Body:          conteudo,
		ContentLength: aws.Int64(tamanho),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("enviando %q ao S3: %w", chave, err)
	}
	return nil
}

func (a *ArmazenamentoS3) Abrir(ctx context.Context, chave string) (io.ReadCloser, error) {
	saida, err := a.cliente.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(chave),
	})
	if err != nil {
		return nil, fmt.Errorf("abrindo %q no S3: %w", chave, err)
	}
	return saida.Body, nil
}

func (a *ArmazenamentoS3) Remover(ctx context.Context, chave string) error {
	_, err := a.cliente.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(chave),
	})
	if err != nil {
		return fmt.Errorf("removendo %q do S3: %w", chave, err)
	}
	return nil
}

// codigoErro informa se err é um erro da API S3 com um dos códigos dados.
func codigoErro(err error, codigos ...string) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	for _, c := range codigos {
		if apiErr.ErrorCode() == c {
			return true
		}
	}
	return false
}
