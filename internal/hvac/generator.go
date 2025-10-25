package hvac

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/patrik-rangel/mock-data-hvac/internal/climate"
)

// HvacSensorData representa os dados sinteticos de um sensor HVAC
type HvacSensorData struct {
	Timestamp            time.Time `json:"timestamp"`            // Momento exato em que os dados foram coletados
	InternalTemperature  float64   `json:"internalTemperature"`  // Temperatura interna medida dentro do espaço (°C)
	SetPointTemperature  float64   `json:"setPointTemperature"`  // Temperatura alvo configurada para o sistema HVAC manter (°C)
	SystemStatus         string    `json:"systemStatus"`         // Estado operacional do sistema: OFF, COOLING, HEATING, FAN_ONLY ou IDLE
	OccupancyStatus      bool      `json:"occupancyStatus"`      // Indica se o espaço está ocupado (true) ou desocupado (false)
	PowerConsumptionKwH  float64   `json:"powerConsumptionKwH"`  // Consumo de energia elétrica do sistema no período (kWh)
	OutdoorTemperature   float64   `json:"outdoorTemperature"`   // Temperatura do ar externo (°C)
	OutdoorHumidity      float64   `json:"outdoorHumidity"`      // Umidade relativa do ar externo (%)
	DeviceId             string    `json:"deviceId"`             // Identificador único do dispositivo ou unidade HVAC (ex: HVAC-UNIT-1)
	SupplyAirTemperature float64   `json:"supplyAirTemperature"` // Temperatura do ar de saída do sistema (°C)
	ReturnAirTemperature float64   `json:"returnAirTemperature"` // Temperatura do ar de retorno para o sistema (°C)
	DuctStaticPressurePa float64   `json:"ductStaticPressurePa"` // Pressão estática nos dutos (Pa)
	CO2LevelPpm          float64   `json:"co2LevelPpm"`          // Nível de CO₂ no ar (ppm)
	RefrigerantPressurePsi float64 `json:"refrigerantPressurePsi"` // Pressão do refrigerante (psi)
	FaultCode            string    `json:"faultCode"`            // Código de falha, se houver
	AssetModel           string    `json:"assetModel"`           // Modelo do equipamento ou ativo
	LocationZone         string    `json:"locationZone"`         // Zona ou localização do dispositivo
}

var (
	randSource rand.Source // Fonte de aleatoriedade
	rng        *rand.Rand  // Gerador de números aleatórios
)

// init inicializa o gerador de números aleatórios global.
func init() {
	randSource = rand.NewSource(time.Now().UnixNano())
	rng = rand.New(randSource)
}

// GenerateHvacData gera um registro de dados HVAC baseado nos dados climáticos e lógica de simulação
func GenerateHvacData(climateData climate.InmetClimateData) HvacSensorData {
	const baseInternalTemp = 22.0
	const setPointDelta = 1.5

	// Simula o estado de saúde e do filtro baseado na época do ano.
	month := climateData.Timestamp.Month()
	floatMonth := float64(month)

	var equipmentHealth float64
	var currentFilterClogLevel float64

	// Simula a manutenção preventiva em Setembro
	if month == time.September {
		equipmentHealth = 0.8 + (rng.Float64() * 0.2) // Saúde: 80% a 100%
		currentFilterClogLevel = rng.Float64() * 0.05 // Entupimento: 0% a 5%
	} else if month > time.September {
		// Período pós-manutenção (Out-Dez)
		equipmentHealth = 0.8 - ((floatMonth - 9.0) / 3.0 * 0.2)
		currentFilterClogLevel = 0.05 + ((floatMonth - 9.0) / 3.0 * 0.4)
	} else {
		// Período pré-manutenção (Jan-Ago)
		equipmentHealth = 1.0 - (floatMonth / 9.0 * 0.4)
		currentFilterClogLevel = floatMonth / 9.0 * 0.8
	}

	// Adiciona variabilidade aleatória e garante limites
	equipmentHealth += (rng.Float64() - 0.5) * 0.1
	equipmentHealth = math.Max(0.4, math.Min(1.0, equipmentHealth))
	currentFilterClogLevel += (rng.Float64() - 0.5) * 0.1
	currentFilterClogLevel = math.Max(0.0, math.Min(1.0, currentFilterClogLevel))

	// Simulação de ocupação
	isOccupied := simulateOccupancy(climateData.Timestamp, rng)
	setPoint := baseInternalTemp + setPointDelta*(rng.Float64()-0.5)

	// Simula a temperatura interna "não controlada"
	uncontrolledInternalTemp := baseInternalTemp + (climateData.TemperatureAir-baseInternalTemp)*0.4 + (rng.Float64()-0.5)*1.5
	internalTempDiff := uncontrolledInternalTemp - setPoint

	// Lógica de decisão do termostato
	systemStatus := "OFF"
	if isOccupied {
		if internalTempDiff > 1.5 {
			systemStatus = "COOLING"
		} else if internalTempDiff < -1.5 {
			systemStatus = "HEATING"
		} else if math.Abs(internalTempDiff) < 1.0 {
			systemStatus = "IDLE"
		}
	}

	// Valores base para as métricas
	supplyTemp := uncontrolledInternalTemp
	ductPressure := 10.0 + rng.Float64()*2.0
	co2Level := 450.0 + (rng.Float64() * 50.0)
	refrigerantPressure := 80.0 + rng.Float64()*5.0
	faultCode := "OK"

	if isOccupied {
		co2Level = 600.0 + (rng.Float64() * 300.0)
		if co2Level > 800.0 && systemStatus == "IDLE" {
			systemStatus = "FAN_ONLY"
		}
	}

	// Simula a temperatura interna final após a ação do HVAC
	finalInternalTemp := uncontrolledInternalTemp
	if systemStatus == "COOLING" {
		finalInternalTemp = setPoint + rng.Float64()*0.5
		supplyTemp = finalInternalTemp - (rng.Float64()*4.0 + 8.0)
		refrigerantPressure = 150.0 + (rng.Float64() * 20.0)
	} else if systemStatus == "HEATING" {
		finalInternalTemp = setPoint - rng.Float64()*0.5
		supplyTemp = finalInternalTemp + (rng.Float64()*3.0 + 5.0)
		refrigerantPressure = 100.0 + (rng.Float64() * 5.0)
	} else if systemStatus == "IDLE" || systemStatus == "FAN_ONLY" {
		finalInternalTemp = setPoint + (rng.Float64()-0.5)*0.5
	}

	// Simulação de falhas
	if systemStatus == "COOLING" && rng.Float64() > equipmentHealth {
		faultCode = "HP-AL-01"
	} else if systemStatus == "HEATING" && rng.Float64() > equipmentHealth {
		faultCode = "HT-FL-02"
	}
	ductPressure += currentFilterClogLevel * 5.0
	if currentFilterClogLevel > 0.8 && rng.Float64() > 0.5 {
		faultCode = "FP-AL-01"
	}
	if ductPressure > 20.0 {
		faultCode = "FP-AL-02"
	}

	// --- AJUSTE FINAL: Simulação de Consumo de Energia (Lógica Aditiva) ---
	powerConsumption := 0.01 // Consumo base (standby)

	// Fator de ineficiência (ADITIVO, não multiplicativo)
	// Saúde (1.0 a 0.4) -> Adiciona 0 a 0.6 kW de custo extra
	// Filtro (0.0 a 1.0) -> Adiciona 0 a 0.4 kW de custo extra
	inefficiencyCost := (1.0 - equipmentHealth)*1.0 + (currentFilterClogLevel * 0.4) // Custo Adicional em kW

	if systemStatus == "COOLING" {
		basePower := 3.0 // AUMENTADO: Potência base (kW) para resfriamento
		// Carga térmica: 400W por grau que a temp. externa excede o setpoint
		tempLoad := math.Max(0, climateData.TemperatureAir-setPoint) * 0.4
		// Carga de umidade: Aumenta consumo drasticamente > 75%
		humidityLoad := 0.0
		if climateData.RelativeHumidity > 75.0 {
			humidityLoad = (climateData.RelativeHumidity - 75.0) / 100.0 * 8.0 // Um pouco menos agressivo
		}
		// Consumo = (Clima) + Ineficiencia Adicional
		powerConsumption = (basePower + tempLoad + humidityLoad) + inefficiencyCost

	} else if systemStatus == "HEATING" {
		basePower := 2.2 // Potência base (kW) para aquecimento (menor que COOLING)
		// Carga térmica: 150W por grau que a temp. externa está abaixo do setpoint
		tempLoad := math.Max(0, setPoint-climateData.TemperatureAir) * 0.15
		// Consumo = (Clima) + Ineficiencia Adicional
		powerConsumption = (basePower + tempLoad) + inefficiencyCost

	} else if systemStatus == "FAN_ONLY" {
		powerConsumption = 0.3 + (rng.Float64() * 0.1) // Consumo da ventoinha (não afetado pela ineficiencia principal)
	}

	powerConsumption *= (1.0 + (rng.Float64()-0.5)*0.1) // Variação aleatória de +/- 5%
	// Garante que o consumo não seja negativo
	powerConsumption = math.Max(0.01, powerConsumption)
	// --- Fim do Bloco de Consumo de Energia ---

	return HvacSensorData{
		Timestamp:              climateData.Timestamp,
		InternalTemperature:    finalInternalTemp,
		SetPointTemperature:    setPoint,
		SystemStatus:           systemStatus,
		OccupancyStatus:        isOccupied,
		PowerConsumptionKwH:    powerConsumption,
		OutdoorTemperature:     climateData.TemperatureAir,
		OutdoorHumidity:        climateData.RelativeHumidity,
		DeviceId:               fmt.Sprintf("SALA-%d", rng.Intn(10)+1),
		SupplyAirTemperature:   supplyTemp,
		ReturnAirTemperature:   finalInternalTemp,
		DuctStaticPressurePa:   ductPressure,
		CO2LevelPpm:            co2Level,
		RefrigerantPressurePsi: refrigerantPressure,
		FaultCode:              faultCode,
		AssetModel:             "HVAC-Model-B",
		LocationZone:           "Zona-A",
	}
}

// simulateOccupancy simula a ocupação baseada no dia da semana e hora.
func simulateOccupancy(t time.Time, r *rand.Rand) bool {
	hour := t.Hour()
	weekday := t.Weekday()

	if hour >= 12 && hour < 14 { // Hora do almoço
		return r.Float64() < 0.30
	}
	if weekday >= time.Monday && weekday <= time.Friday {
		if hour >= 8 && hour < 18 { // Horário comercial
			return r.Float64() < 0.90
		}
	}
	return r.Float64() < 0.10 // Fora do horário comercial / Fim de semana
}

// WriteJSON converte um slice de HvacSensorData para JSON formatado.
func WriteJSON(data []HvacSensorData) ([]byte, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ") // Usa 2 espaços para indentação
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar dados HVAC para JSON: %w", err)
	}
	return jsonData, nil
}