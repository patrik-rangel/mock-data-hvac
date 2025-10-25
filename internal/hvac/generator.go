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
	// Isso evita a degradação linear contínua que ocorria com variáveis globais.
	month := climateData.Timestamp.Month()
	floatMonth := float64(month)

	var equipmentHealth float64
	var currentFilterClogLevel float64

	// Simula a manutenção preventiva em Setembro
	if month == time.September {
		// Saúde melhora, filtro é limpo
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
	equipmentHealth = math.Max(0.4, math.Min(1.0, equipmentHealth)) // Limite de saúde 40%-100%
	currentFilterClogLevel += (rng.Float64() - 0.5) * 0.1
	currentFilterClogLevel = math.Max(0.0, math.Min(1.0, currentFilterClogLevel))

	// Simulação de ocupação
	isOccupied := simulateOccupancy(climateData.Timestamp, rng)
	setPoint := baseInternalTemp + setPointDelta*(rng.Float64()-0.5)

	// Simula a temperatura interna "não controlada" (inércia térmica)
	uncontrolledInternalTemp := baseInternalTemp + (climateData.TemperatureAir-baseInternalTemp)*0.4 + (rng.Float64()-0.5)*1.5
	// Diferença entre a temp. interna (sem HVAC) e o setpoint
	internalTempDiff := uncontrolledInternalTemp - setPoint

	// Lógica de decisão do termostato
	systemStatus := "OFF"
	if isOccupied {
		if internalTempDiff > 1.5 { // Se 1.5°C acima do setpoint
			systemStatus = "COOLING"
		} else if internalTempDiff < -1.5 { // Se 1.5°C abaixo do setpoint
			systemStatus = "HEATING"
		} else if math.Abs(internalTempDiff) < 1.0 { // Dentro da "banda morta"
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
		if co2Level > 800.0 && systemStatus == "IDLE" { // Lógica de ventilação (qualidade do ar)
			systemStatus = "FAN_ONLY"
		}
	}

	// Simula a temperatura interna final após a ação do HVAC
	finalInternalTemp := uncontrolledInternalTemp
	if systemStatus == "COOLING" {
		finalInternalTemp = setPoint + rng.Float64()*0.5 // Temp. estabiliza próximo ao setpoint
		supplyTemp = finalInternalTemp - (rng.Float64()*4.0 + 8.0) // Ar de insuflamento frio
		refrigerantPressure = 150.0 + (rng.Float64() * 20.0)
	} else if systemStatus == "HEATING" {
		finalInternalTemp = setPoint - rng.Float64()*0.5 // Temp. estabiliza próximo ao setpoint
		supplyTemp = finalInternalTemp + (rng.Float64()*3.0 + 5.0) // Ar de insuflamento quente
		refrigerantPressure = 100.0 + (rng.Float64() * 5.0)
	} else if systemStatus == "IDLE" || systemStatus == "FAN_ONLY" {
		finalInternalTemp = setPoint + (rng.Float64()-0.5)*0.5 // Temp. flutua próximo ao setpoint
	}
	// Se systemStatus == "OFF", a temp. interna é a uncontrolledInternalTemp

	// Simulação de falhas baseada na saúde e no filtro
	if systemStatus == "COOLING" && rng.Float64() > equipmentHealth {
		faultCode = "HP-AL-01" // Alarme de alta pressão
	} else if systemStatus == "HEATING" && rng.Float64() > equipmentHealth {
		faultCode = "HT-FL-02" // Falha no aquecimento
	}

	ductPressure += currentFilterClogLevel * 5.0 // Filtro sujo aumenta a pressão do duto
	if currentFilterClogLevel > 0.8 && rng.Float64() > 0.5 {
		faultCode = "FP-AL-01" // Alarme de filtro entupido
	}
	if ductPressure > 20.0 {
		faultCode = "FP-AL-02" // Alarme de alta pressão no duto
	}

	// Simulação de Consumo de Energia
	powerConsumption := 0.01 // Consumo base (standby)

	// Fator de ineficiência: Saúde ruim e filtro sujo aumentam o consumo
	// Saúde (1.0 a 0.4) -> Adiciona 0% a 30% de custo
	// Filtro (0.0 a 1.0) -> Adiciona 0% a 25% de custo
	inefficiencyFactor := 1.0 + (1.0 - equipmentHealth)*0.5 + (currentFilterClogLevel * 0.25)

	if systemStatus == "COOLING" {
		basePower := 2.0 // Potência base (kW) para resfriamento
		// Carga térmica: 500W por grau que a temp. externa excede o setpoint
		tempLoad := math.Max(0, climateData.TemperatureAir-setPoint) * 0.5
		// Carga de umidade: Aumenta consumo drasticamente > 75% (desumidificação)
		humidityLoad := 0.0
		if climateData.RelativeHumidity > 75.0 {
			humidityLoad = (climateData.RelativeHumidity - 75.0) / 100.0 * 10.0
		}
		powerConsumption = (basePower + tempLoad + humidityLoad) * inefficiencyFactor

	} else if systemStatus == "HEATING" {
		basePower := 2.5 // Potência base (kW) para aquecimento (resistivo)
		// Carga térmica: 200W por grau que a temp. externa está abaixo do setpoint
		tempLoad := math.Max(0, setPoint-climateData.TemperatureAir) * 0.2
		powerConsumption = (basePower + tempLoad) * inefficiencyFactor

	} else if systemStatus == "FAN_ONLY" {
		powerConsumption = 0.4 + (rng.Float64() * 0.1) // Consumo da ventoinha
	}

	powerConsumption *= (1.0 + (rng.Float64()-0.5)*0.1) // Variação aleatória de +/- 5%

	return HvacSensorData{
		Timestamp:              climateData.Timestamp,
		InternalTemperature:    finalInternalTemp,
		SetPointTemperature:    setPoint,
		SystemStatus:           systemStatus,
		OccupancyStatus:        isOccupied,
		PowerConsumptionKwH:    powerConsumption,
		OutdoorTemperature:     climateData.TemperatureAir,
		OutdoorHumidity:        climateData.RelativeHumidity,
		DeviceId:               fmt.Sprintf("SALA-%d", rng.Intn(10)+1), // ID aleatório (1-10)
		SupplyAirTemperature:   supplyTemp,
		ReturnAirTemperature:   finalInternalTemp, // Temp. de retorno é a temp. interna
		DuctStaticPressurePa:   ductPressure,
		CO2LevelPpm:            co2Level,
		RefrigerantPressurePsi: refrigerantPressure,
		FaultCode:              faultCode,
		AssetModel:             "HVAC-Model-B",
		LocationZone:           "Zona-A",
	}
}

// simulateOccupancy simula a ocupação baseada no dia da semana e hora.
// Inclui probabilidade reduzida no horário de almoço.
func simulateOccupancy(t time.Time, r *rand.Rand) bool {
	hour := t.Hour()
	weekday := t.Weekday()

	// Simula a hora do almoço (ocupação cai)
	if hour >= 12 && hour < 14 {
		return r.Float64() < 0.30 // 30% de chance
	}

	if weekday >= time.Monday && weekday <= time.Friday {
		if hour >= 8 && hour < 18 { // Horário comercial (8h-12h e 14h-18h)
			return r.Float64() < 0.90 // 90% de chance
		}
	}
	// Fins de semana e fora do horário comercial
	return r.Float64() < 0.10 // 10% de chance
}

// WriteJSON converte um slice de HvacSensorData para JSON formatado.
func WriteJSON(data []HvacSensorData) ([]byte, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar dados HVAC para JSON: %w", err)
	}
	return jsonData, nil
}