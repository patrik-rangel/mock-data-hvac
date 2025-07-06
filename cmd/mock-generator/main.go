package main

import (
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"

	"github.com/patrik-rangel/mock-data-hvac/internal/climate"
	"github.com/patrik-rangel/mock-data-hvac/internal/hvac"
)

func main() {
	// 1. Carregar variáveis de ambiente (ainda útil para outras configs futuras, mas S3_BUCKET_NAME e AWS_REGION não serão usados aqui)
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: Não foi possível carregar o arquivo .env. Erro:", err)
	}

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

	// 7. Definir o nome do arquivo JSON local
	localFileName := fmt.Sprintf("hvac_mock_data_A701_%s.json", time.Now().Format("20060102_150405"))

	fmt.Printf("Salvando dados JSON localmente como: %s\n", localFileName)

	// 8. Salvar o JSON em um arquivo local usando a nova função do pacote 'hvac'
	err = hvac.SaveJSONLocally(jsonData, localFileName)
	if err != nil {
		log.Fatalf("Erro fatal ao salvar o JSON localmente: %v", err)
	}

	fmt.Println("Processo concluído com sucesso! Dados mocados salvos localmente.")
}
