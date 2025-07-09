package hvac

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/patrik-rangel/mock-data-hvac/internal/climate"
)

// HvacSensorData representa os dados mocados de um sensor HVAC
type HvacSensorData struct {
    // Timestamp: O momento exato em que os dados do sensor foram "coletados".
    // Formato ISO 8601 (e.g., "2024-01-01T15:00:00Z").
    Timestamp time.Time `json:"timestamp"`

    // InternalTemperature: A temperatura ambiente medida dentro do espaço
    // onde o sensor HVAC está localizado, em graus Celsius.
    InternalTemperature float64 `json:"internalTemperature"`

    // SetPointTemperature: A temperatura alvo que o sistema HVAC está configurado
    // para manter, definida pelo usuário ou por um algoritmo de controle, em graus Celsius.
    SetPointTemperature float64 `json:"setPointTemperature"`

    // SystemStatus: O estado operacional atual do sistema HVAC.
    // Exemplos de valores: "OFF", "COOLING" (resfriando), "HEATING" (aquecendo),
    // "FAN_ONLY" (apenas ventilação), "IDLE" (ocioso).
    SystemStatus string `json:"systemStatus"`

    // OccupancyStatus: Indica se o espaço onde o sensor está localizado está ocupado
    // no momento da coleta dos dados (true) ou desocupado (false).
    // Fundamental para correlacionar consumo e comportamento com a presença humana.
    OccupancyStatus bool `json:"occupancyStatus"`

    // PowerConsumptionKwH: O consumo de energia elétrica do sistema HVAC
    // no período de amostragem, em Quilowatt-hora (kWh).
    // Este valor é simulado com base no status do sistema, temperatura externa e ocupação.
    PowerConsumptionKwH float64 `json:"powerConsumptionKwH"`

    // OutdoorTemperature: A temperatura do ar externo no momento da coleta dos dados,
    // em graus Celsius. Estes dados são provenientes da estação INMET.
    // Crucial para entender a demanda do sistema HVAC.
    OutdoorTemperature float64 `json:"outdoorTemperature"`

    // OutdoorHumidity: A umidade relativa do ar externo no momento da coleta dos dados,
    // em porcentagem (%). Estes dados são provenientes da estação INMET.
    // Também influencia a carga térmica e o conforto.
    OutdoorHumidity float64 `json:"outdoorHumidity"`

    // DeviceId: Um identificador único para o dispositivo ou unidade HVAC que gerou os dados.
    // Permite distinguir dados de diferentes sistemas ou localizações.
    // Exemplo: "HVAC-UNIT-1", "HVAC-UNIT-A701-001".
    DeviceId string `json:"deviceId"`
}

// Criar uma variável global para o gerador de números aleatórios.
// randSource será o "semeador" e rng será o gerador.
var (
	randSource rand.Source
	rng        *rand.Rand
)

// init é uma função especial que é executada uma vez quando o pacote é inicializado.
// Usamos ela para semear o gerador de números aleatórios de forma segura.
func init() {
	randSource = rand.NewSource(time.Now().UnixNano())
	rng = rand.New(randSource)
}

// GenerateHvacData gera um registro de dados HVAC baseado nos dados climáticos e lógica de simulação
func GenerateHvacData(climateData climate.InmetClimateData) HvacSensorData {
	const baseInternalTemp = 23.0
	const setPointDelta = 2.0
	const baseConsumption = 0.5
	const coolingConsumptionRate = 0.2
	const heatingConsumptionRate = 0.1

	// Simula ocupação usando o gerador 'rng'
	isOccupied := simulateOccupancy(climateData.Timestamp, rng)

	internalTemp := baseInternalTemp + (rng.Float64()-0.5)*1.5
	setPoint := baseInternalTemp + setPointDelta*(rng.Float64()-0.5)

	systemStatus := "OFF"
	powerConsumption := baseConsumption

	if isOccupied {
		if climateData.TemperatureAir > (setPoint + 1.0) {
			systemStatus = "COOLING"
			powerConsumption += coolingConsumptionRate * (climateData.TemperatureAir - setPoint) * (rng.Float64() * 2) // Usar rng.Float64()
		} else if climateData.TemperatureAir < (setPoint - 3.0) {
			systemStatus = "HEATING"
			powerConsumption += heatingConsumptionRate * (setPoint - climateData.TemperatureAir) * (rng.Float64() * 1.5) // Usar rng.Float64()
		} else {
			systemStatus = "FAN_ONLY"
			powerConsumption += baseConsumption * (rng.Float64() * 0.5)
		}
	} else {
		systemStatus = "OFF"
		powerConsumption = baseConsumption * rng.Float64() * 0.1
	}

	if powerConsumption < 0 {
		powerConsumption = 0
	}

	return HvacSensorData{
		Timestamp:           climateData.Timestamp,
		InternalTemperature: internalTemp,
		SetPointTemperature: setPoint,
		SystemStatus:        systemStatus,
		OccupancyStatus:     isOccupied,
		PowerConsumptionKwH: powerConsumption,
		OutdoorTemperature:  climateData.TemperatureAir,
		OutdoorHumidity:     climateData.RelativeHumidity,
		DeviceId:            fmt.Sprintf("HVAC-UNIT-%d", rng.Intn(10)+1), // Usar rng.Intn()
	}
}

// simulateOccupancy simula a ocupação com base na hora do dia
// Agora aceita um gerador *rand.Rand
func simulateOccupancy(t time.Time, r *rand.Rand) bool {
	hour := t.Hour()
	weekday := t.Weekday()

	if weekday >= time.Monday && weekday <= time.Friday {
		if hour >= 8 && hour < 18 {
			return r.Float64() < 0.90
		}
	}
	return r.Float64() < 0.10
}

func WriteJSON(data []HvacSensorData) ([]byte, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar dados HVAC para JSON: %w", err)
	}
	return jsonData, nil
}
