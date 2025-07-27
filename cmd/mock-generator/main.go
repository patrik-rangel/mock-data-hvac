package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/patrik-rangel/mock-data-hvac/internal/climate"
	"github.com/patrik-rangel/mock-data-hvac/internal/hvac"
	"github.com/patrik-rangel/mock-data-hvac/internal/s3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: Não foi possível carregar o arquivo .env. Erro:", err)
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")
	awsRegion := os.Getenv("AWS_REGION")
	endpointUrl := os.Getenv("ENDPOINT_URL")

	// 3. Definir o caminho do arquivo CSV do INMET
	inmetCSVPath := "data/inmet/dados-202401-202501.zip"
	fmt.Printf("Lendo dados climáticos do CSV: %s\n", inmetCSVPath)

	// 4. Ler os dados climáticos do CSV
	climateRecords, err := climate.ReadInmetCSV(inmetCSVPath)
	if err != nil {
		log.Fatalf("Erro fatal ao ler dados do INMET: %v", err)
	}
	fmt.Printf("Lidos %d registros climáticos do INMET.\n", len(climateRecords))

	if len(climateRecords) == 0 {
		log.Println("Nenhum registro climático encontrado no CSV. Saindo.")
		return
	}

	// 5. Gerar os dados de sensores HVAC mocados
	fmt.Println("Iniciando a geração de dados de sensores HVAC mocados...")

	var allHvacData []hvac.HvacSensorData
	for _, record := range climateRecords {
		hvacData := hvac.GenerateHvacData(record)
		allHvacData = append(allHvacData, hvacData)
	}
	fmt.Printf("Gerados %d registros de dados HVAC mocados.\n", len(allHvacData))

	// 6. Converter os dados HVAC mocados para JSON
	fmt.Println("Convertendo dados HVAC para formato JSON...")
	jsonData, err := hvac.WriteJSON(allHvacData)
	if err != nil {
		log.Fatalf("Erro fatal ao converter dados HVAC para JSON: %v", err)
	}
	fmt.Println("Dados HVAC convertidos para JSON com sucesso.")

	// 7. Definir o nome do arquivo JSON no bucket
	localFileName := fmt.Sprintf("hvac_mock_data_%s.json", time.Now())

	fmt.Printf("Salvando dados JSON no bucket como: %s\n", localFileName)

	err = s3.UploadDataToS3(bucketName, awsRegion, endpointUrl, jsonData, localFileName)
	if err != nil {
		log.Fatalf("Erro fatal ao salvar o JSON no bucket: %v", err)
	}

	fmt.Println("Processo concluído com sucesso! Dados mocados salvos no s3.")
}
