package hvac

import (
	"encoding/json"
	"fmt"
	"os"
)

func WriteHvacDataToJSONL(filename string, data []HvacSensorData) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range data {
		if err := encoder.Encode(record); err != nil {
			return err
		}
	}
	return nil
}

func SaveJSONLocally(jsonData []byte, filename string) error {
	err := os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("erro ao salvar o arquivo JSON localmente '%s': %w", filename, err)
	}
	return nil
}
