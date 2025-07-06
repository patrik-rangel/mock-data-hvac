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

func UploadDataToS3(bucketName, region string, data []byte, key string) error {
	log.Printf("Iniciando upload de '%s' para o bucket S3 '%s' na região '%s'...", key, bucketName, region)

	// Carrega a configuração padrão da AWS.
	// O SDK buscará credenciais nas variáveis de ambiente (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
	// ou no arquivo de credenciais (~/.aws/credentials).
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("falha ao carregar a configuração AWS: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	putObjectInput := &s3.PutObjectInput{
		Bucket: aws.String(bucketName), // Nome do bucket
		Key:    aws.String(key),        // Caminho/nome do arquivo no S3
		Body:   bytes.NewReader(data),  // Os dados em bytes
		// ContentType: aws.String("application/json"), // Opcional: Define o tipo de conteúdo
	}

	_, err = client.PutObject(context.TODO(), putObjectInput)
	if err != nil {
		return fmt.Errorf("falha ao fazer upload para S3: %w", err)
	}

	log.Printf("Upload de '%s' para S3 concluído com sucesso!", key)
	return nil
}
