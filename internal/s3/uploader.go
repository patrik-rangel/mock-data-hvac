package s3

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadDataToS3(bucketName, region, awsEndpointURL string, data []byte, key string) error {
	log.Printf("Iniciando upload de '%s' para o bucket S3 '%s' na região '%s'...", key, bucketName, region)

	// Opções de configuração para o SDK
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	// Se um endpoint customizado for fornecido (para LocalStack)
	if awsEndpointURL != "" {
		log.Printf("Usando endpoint S3 customizado: %s\n", awsEndpointURL)
		opts = append(opts, config.WithBaseEndpoint(awsEndpointURL))
	}

	// Carrega a configuração padrão da AWS com as opções
	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return fmt.Errorf("falha ao carregar a configuração AWS: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	putObjectInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	}

	_, err = client.PutObject(context.TODO(), putObjectInput)
	if err != nil {
		return fmt.Errorf("falha ao fazer upload para S3: %w", err)
	}

	log.Printf("Upload de '%s' para S3 concluído com sucesso!", key)
	return nil
}
