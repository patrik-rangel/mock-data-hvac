package climate

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type InmetClimateData struct {
	Timestamp        time.Time
	TemperatureAir   float64
	RelativeHumidity float64
}

func ReadInmetCSV(filepath string) ([]InmetClimateData, error) {
	var reader io.Reader
	var closer io.Closer

	ext := strings.ToLower(filepath[strings.LastIndex(filepath, ".")+1:])

	if ext == "zip" {
		zipReader, err := zip.OpenReader(filepath)
		if err != nil {
			return nil, fmt.Errorf("erro ao abrir arquivo ZIP '%s': %w", filepath, err)
		}
		closer = zipReader

		var csvFile *zip.File
		for _, f := range zipReader.File {
			if strings.HasSuffix(strings.ToLower(f.Name), ".csv") {
				csvFile = f
				break
			}
		}

		if csvFile == nil {
			return nil, fmt.Errorf("nenhum arquivo CSV encontrado dentro do ZIP '%s'", filepath)
		}

		rc, err := csvFile.Open()
		if err != nil {
			return nil, fmt.Errorf("erro ao abrir arquivo CSV dentro do ZIP '%s': %w", csvFile.Name, err)
		}
		reader = rc
		defer rc.Close()
	} else if ext == "csv" {
		file, err := os.Open(filepath)
		if err != nil {
			return nil, fmt.Errorf("erro ao abrir o arquivo CSV '%s': %w", filepath, err)
		}
		reader = file
		closer = file
	} else {
		return nil, fmt.Errorf("formato de arquivo não suportado: '%s'. Esperado .csv ou .zip", ext)
	}

	if closer != nil {
		defer func() {
			if cerr := closer.Close(); cerr != nil {
				log.Printf("Aviso: Erro ao fechar o leitor: %v", cerr)
			}
		}()
	}

	csvReader := csv.NewReader(reader)
	csvReader.Comma = ';'
	csvReader.FieldsPerRecord = -1

	var climateData []InmetClimateData
	headerMap := make(map[string]int)
	headerFound := false

	for i := 0; ; i++ {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("erro ao ler linha %d do CSV: %w", i+1, err)
		}

		if i < 9 {
			continue
		} else if i == 9 {
			for idx, colName := range record {
				normalizedColName := strings.TrimSpace(strings.ToLower(colName))
				if strings.Contains(normalizedColName, "(") {
					normalizedColName = strings.Split(normalizedColName, "(")[0]
					normalizedColName = strings.TrimSpace(normalizedColName)
				}
				headerMap[normalizedColName] = idx
			}
			headerFound = true

			_, hasDate := headerMap["data medicao"]
			_, hasTime := headerMap["hora medicao"]
			_, hasTemp := headerMap["temperatura do ar - bulbo seco, horaria"]
			_, hasHum := headerMap["umidade relativa do ar, horaria"]

			if !hasDate || !hasTime || !hasTemp || !hasHum {
				return nil, fmt.Errorf("cabeçalho do CSV não contém todas as colunas esperadas. Verifique os nomes das colunas no CSV e no código: %v", headerMap)
			}
			continue
		}

		if !headerFound {
			return nil, fmt.Errorf("dados encontrados antes do cabeçalho ser mapeado. Verifique a estrutura do CSV. %v", headerFound)
		}

		dateStr := record[headerMap["data medicao"]]
		timeStrRaw := record[headerMap["hora medicao"]]

		var timeStrFormatted string
		if len(timeStrRaw) == 4 {
			timeStrFormatted = timeStrRaw[:2] + ":" + timeStrRaw[2:]
		} else if len(timeStrRaw) == 3 {
			timeStrFormatted = "0" + timeStrRaw[:1] + ":" + timeStrRaw[1:]
		} else {
			log.Printf("Aviso: Formato de hora inesperado '%s' na linha %d. Pulando linha.", timeStrRaw, i+1)
			continue
		}

		dateTimeStr := fmt.Sprintf("%s %s", dateStr, timeStrFormatted)
		timestamp, err := time.Parse("2006-01-02 15:04", dateTimeStr)
		if err != nil {
			log.Printf("Aviso: Erro ao fazer parse do timestamp '%s' na linha %d: %v. Pulando linha.", dateTimeStr, i+1, err)
			continue
		}

		tempAirStr := strings.Replace(record[headerMap["temperatura do ar - bulbo seco, horaria"]], ",", ".", -1)
		tempAir, err := strconv.ParseFloat(tempAirStr, 64)
		if err != nil {
			log.Printf("Aviso: Erro ao fazer parse da temperatura do ar '%s' na linha %d: %v. Pulando linha.", tempAirStr, i+1, err)
			continue
		}

		humidityStr := strings.Replace(record[headerMap["umidade relativa do ar, horaria"]], ",", ".", -1)
		humidity, err := strconv.ParseFloat(humidityStr, 64)
		if err != nil {
			log.Printf("Aviso: Erro ao fazer parse da umidade relativa '%s' na linha %d: %v. Pulando linha.", humidityStr, i+1, err)
			continue
		}

		climateData = append(climateData, InmetClimateData{
			Timestamp:        timestamp,
			TemperatureAir:   tempAir,
			RelativeHumidity: humidity,
		})
	}

	return climateData, nil
}
