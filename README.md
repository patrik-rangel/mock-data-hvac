# HVAC Mock Data Generator 🌡️

O **mock-data-hvac** é um simulador de dados sintéticos inteligente de sistemas de climatização e ventilação. Diferente de geradores puramente aleatórios, este serviço utiliza **dados reais do clima de São Paulo** (estação A701 do INMET) e lógica aplicada para reproduzir o comportamento real de uma máquina de HVAC ao longo de um ano.

---

## 🏗️ Como o sistema funciona?

O simulador transforma variáveis climáticas brutas em telemetria detalhada de sensores:

1.  **Leitura de Dados Reais:** Consome arquivos horários de temperatura e umidade do INMET (período 2024-2025).
2.  **Simulação de Conforto Térmico:** Calcula como a temperatura interna reage ao clima externo e à atuação do ar-condicionado, considerando a inércia térmica do ambiente.
3.  **Cálculo de Consumo de Energia:** Estima o gasto em $kWh$ com base na diferença de temperatura e no esforço necessário para retirar a umidade excessiva do ar.
4.  **Ciclo de Vida do Ativo:** Modela o desgaste natural do equipamento, simulando a perda de saúde do compressor e o entupimento progressivo dos filtros.
5.  **Sincronização Cloud:** Consolida os resultados em um JSON estruturado e realiza o upload automático para o **AWS S3**.



---

## 🔬 Lógica da Simulação

Para garantir que os dashboards de BI reflitam cenários do mundo real, o simulador utiliza regras fundamentadas:

### 1. Inércia Térmica
A temperatura dentro de uma sala não muda instantaneamente. O sistema simula a resistência térmica do edifício, onde a temperatura interna tenta se equilibrar com a externa, enquanto o HVAC atua para corrigi-la.

### 2. Impacto da Umidade no Consumo
Remover umidade (calor latente) exige muito mais energia do que apenas baixar a temperatura. Quando a umidade relativa ultrapassa **75%**, o simulador aumenta o consumo para representar o esforço de desumidificação



---

## 🚀 Principais Funcionalidades

* **Códigos de Falha Reais:** Gera alarmes técnicos como `HP-AL-01` (Alta Pressão) e `FP-AL-01` (Filtro Sujo) baseados no desgaste da máquina.
* **Sazonalidade de Manutenção:** Simula a degradação do filtro ao longo dos meses e uma janela de manutenção preventiva em Setembro, onde os parâmetros de eficiência são resetados.
* **Ocupação Dinâmica:** Probabilidade de presença humana baseada em dias úteis e horários comerciais, incluindo a variação do nível de CO_2.

---

## 🛠️ Tecnologias

* **Linguagem:** Go (Golang)
* **Dados:** INMET (São Paulo - 2024/2025)
* **Nuvem:** AWS SDK for Go v2 (S3)

---

## ⚙️ Configuração e Execução

1.  Configure as variáveis de acesso ao S3 no arquivo `.env`:
    ```env
    S3_BUCKET_NAME=seu-bucket
    AWS_REGION=us-east-1
    ENDPOINT_URL=http://localhost:4566
    ```
2.  Certifique-se de que os dados do INMET estão em `data/inmet/`.
3.  Instale as dependências e rode o serviço:
    ```bash
    go mod tidy
    go run main.go
    ```

---
