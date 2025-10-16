package hvac

import (
	"encoding/json"
	"fmt"
	"math"
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

	SupplyAirTemperature   float64 `json:"supplyAirTemperature"`   // Temperatura do ar de saída do sistema
	ReturnAirTemperature   float64 `json:"returnAirTemperature"`   // Temperatura do ar que retorna para o sistema
	DuctStaticPressurePa   float64 `json:"ductStaticPressurePa"`   // Pressão estática nos dutos, em Pascal (Pa)
	CO2LevelPpm            float64 `json:"co2LevelPpm"`            // Nível de CO2 no ar, em partes por milhão (ppm)
	RefrigerantPressurePsi float64 `json:"refrigerantPressurePsi"` // Pressão do refrigerante, em psi
	FaultCode              string  `json:"faultCode"`              // Código de falha, se houver
	AssetModel             string  `json:"assetModel"`
	LocationZone           string  `json:"locationZone"`
}

// Criar uma variável global para o gerador de números aleatórios.
// randSource será o "semeador" e rng será o gerador.
var (
	randSource             rand.Source
	rng                    *rand.Rand
	currentFilterClogLevel float64
	equipmentHealth        float64
)

// init é uma função especial que é executada uma vez quando o pacote é inicializado.
// Usamos ela para semear o gerador de números aleatórios de forma segura.
func init() {
	randSource = rand.NewSource(time.Now().UnixNano())
	rng = rand.New(randSource)
	currentFilterClogLevel = 0.0 // O filtro começa limpo
	equipmentHealth = 1.0        // O equipamento começa 100% saudável
}

// GenerateHvacData gera um registro de dados HVAC baseado nos dados climáticos e lógica de simulação
func GenerateHvacData(climateData climate.InmetClimateData) HvacSensorData {
	const baseInternalTemp = 22.0
	const setPointDelta = 2.0

	// Lógica de degradação da saúde do equipamento e entupimento do filtro.
	// A degradação da saúde do equipamento é contínua e não se recupera.
	equipmentHealth -= rng.Float64() * 0.0001
	if equipmentHealth < 0.2 { // Saúde mínima para evitar valores absurdos.
		equipmentHealth = 0.2
	}

	// A sujeira do filtro aumenta com o tempo.
	currentFilterClogLevel += rng.Float64() * 0.001
	if currentFilterClogLevel > 1.0 { // Limite máximo para a sujeira do filtro.
		currentFilterClogLevel = 1.0
	}

	// Simulação da manutenção preventiva: filtro é limpo em setembro.
	if climateData.Timestamp.Month() == time.September {
		currentFilterClogLevel = rng.Float64() * 0.05 // O filtro é limpo no início de setembro.
	}

	// Simulação de ocupação.
	isOccupied := simulateOccupancy(climateData.Timestamp, rng)

	// Definição de valores base para as métricas.
	returnTemp := baseInternalTemp + (climateData.TemperatureAir-baseInternalTemp)*0.25 + (rng.Float64()-0.5)*1.5
	setPoint := baseInternalTemp + setPointDelta*(rng.Float64()-0.5)
	supplyTemp := returnTemp
	ductPressure := 10.0 + rng.Float64()*2.0
	co2Level := 450.0 + (rng.Float64() * 50.0)
	refrigerantPressure := 80.0 + rng.Float64()*5.0
	faultCode := "OK"
	systemStatus := "OFF"
	powerConsumption := 0.0

	if isOccupied {
		co2Level = 600.0 + (rng.Float64() * 300.0) // CO2 aumenta com a ocupação.
		if co2Level > 800.0 && systemStatus == "OFF" {
			systemStatus = "FAN_ONLY"
		}
	}

	// Lógica de controle do sistema HVAC baseada na ocupação e temperatura externa.
	tempDiff := climateData.TemperatureAir - setPoint
	if isOccupied {
		if math.Abs(tempDiff) > 4.0 { // Diferença grande, consome muito.
			if tempDiff > 0 { // Se a temperatura externa está mais alta.
				systemStatus = "COOLING"
				supplyTemp = returnTemp - (rng.Float64()*4.0 + 8.0) // Delta-T de 8 a 12 graus.
				refrigerantPressure = 150.0 + (rng.Float64() * 20.0)
			} else { // Se a temperatura externa está mais baixa.
				systemStatus = "HEATING"
				supplyTemp = returnTemp + (rng.Float64()*3.0 + 5.0) // Delta-T de 5 a 8 graus.
				refrigerantPressure = 100.0 + (rng.Float64() * 5.0)
			}
		} else if math.Abs(tempDiff) > 1.5 { // Diferença moderada.
			systemStatus = "FAN_ONLY"
			ductPressure = 12.0 + rng.Float64()
		} else { // Diferença pequena, sistema pode ficar em OFF.
			systemStatus = "OFF"
		}
	} else {
		systemStatus = "OFF"
	}

	// Lógica para gerar falhas baseadas na saúde do equipamento e no entupimento do filtro.
	if systemStatus == "COOLING" {
		if rng.Float64() > equipmentHealth {
			faultCode = "HP-AL-01" // High Pressure Alarm.
		}
	} else if systemStatus == "HEATING" {
		if rng.Float64() > equipmentHealth {
			faultCode = "HT-FL-02" // Heating Failure.
		}
	}

	if currentFilterClogLevel > 0.6 && rng.Float64() > 0.5 {
		faultCode = "FP-AL-01"
	}

	// A pressão do duto aumenta com o entupimento do filtro.
	ductPressure += currentFilterClogLevel * 5.0
	if ductPressure > 20.0 {
		faultCode = "FP-AL-02" // Filter Pressure Alarm.
	}

	// Lógica para simular o consumo de energia.
	if systemStatus == "COOLING" {
		powerConsumption = (2.0 + tempDiff*4.0 + (climateData.RelativeHumidity/100.0)*5.0) * (2.0 - equipmentHealth)
		powerConsumption *= (1.0 + currentFilterClogLevel*0.5)
	} else if systemStatus == "HEATING" {
		powerConsumption = (2.0 + math.Abs(tempDiff)*1.5) * (2.0 - equipmentHealth)
	} else if systemStatus == "FAN_ONLY" {
		powerConsumption = 0.3 + rng.Float64()*0.1
	}

	return HvacSensorData{
		Timestamp:              climateData.Timestamp,
		InternalTemperature:    returnTemp,
		SetPointTemperature:    setPoint,
		SystemStatus:           systemStatus,
		OccupancyStatus:        isOccupied,
		PowerConsumptionKwH:    powerConsumption,
		OutdoorTemperature:     climateData.TemperatureAir,
		OutdoorHumidity:        climateData.RelativeHumidity,
		DeviceId:               fmt.Sprintf("SALA-%d", rng.Intn(10)+1),
		SupplyAirTemperature:   supplyTemp,
		ReturnAirTemperature:   returnTemp,
		DuctStaticPressurePa:   ductPressure,
		CO2LevelPpm:            co2Level,
		RefrigerantPressurePsi: refrigerantPressure,
		FaultCode:              faultCode,
		AssetModel:             "HVAC-Model-B",
		LocationZone:           "Zona-A",
	}
}

// simulateOccupancy simula a ocupação com base na hora do dia
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
