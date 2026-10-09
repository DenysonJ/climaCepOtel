# Clima por CEP com Observabilidade

Sistema distribuído em Go com dois microsserviços que consultam o clima de uma cidade a partir de um CEP, instrumentados com **OpenTelemetry** e **Zipkin** para rastreamento distribuído.

- **Serviço A (`cep`)** — porta `8080`: recebe o CEP via POST, valida e encaminha para o Serviço B.
- **Serviço B (`clima`)** — porta `8181`: consulta a cidade (ViaCEP) e a temperatura (WeatherAPI) e devolve as conversões.
- **OTEL Collector** + **Zipkin**: coleta e visualização dos traços.

## Pré-requisitos

- Docker e Docker Compose.
- Uma chave da [WeatherAPI](https://www.weatherapi.com/).

Crie um arquivo `.env` na raiz (baseado no `.env.example`) com a sua chave:

```bash
WEB_SERVER_PORT=8080
WEATHER_API_KEY=sua_chave_aqui
```

## Como subir o ambiente

Sobe os dois serviços, o OTEL Collector e o Zipkin:

```bash
docker-compose up --build
```

## Como fazer a requisição (Serviço A)

O Serviço A aceita `POST /cep` na porta `8080`, com o CEP (string de 8 dígitos) no corpo JSON.

```bash
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep":"29902555"}'
```

### Resposta de sucesso — `200 OK`

```json
{
  "city": "São Paulo",
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.65
}
```

### Erros

| Status | Mensagem               | Quando ocorre                                  |
|--------|------------------------|------------------------------------------------|
| 422    | `invalid zipcode`      | CEP com formato inválido (não tem 8 dígitos)   |
| 404    | `can not find zipcode` | CEP com formato correto, mas não encontrado    |

Exemplo de CEP inválido (retorna `422`):

```bash
curl -i -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep":"123"}'
```

## Como visualizar os traços no Zipkin

Com o ambiente no ar, faça ao menos uma requisição ao Serviço A e abra o Zipkin no navegador:

```
http://localhost:9411
```

Clique em **Run Query** (ou filtre por `serviceName = microservice_cep`) para ver o fluxo completo da requisição:

```
POST /cep (Serviço A) → GET /clima (Serviço B) → Busca de CEP (ViaCEP) → Busca de temperatura (WeatherAPI)
```

Cada traço exibe os spans manuais de **busca de CEP** e **busca de temperatura**, com seus respectivos tempos de resposta.

## Portas expostas

| Serviço           | Porta  | Descrição                       |
|-------------------|--------|---------------------------------|
| Serviço A (cep)   | 8080   | Porta de entrada (POST /cep)    |
| Serviço B (clima) | 8181   | Orquestração (GET /clima?cep=)  |
| Zipkin            | 9411   | UI de rastreamento distribuído  |
| OTEL Collector    | 4317   | Recebe os traços via OTLP/gRPC  |

## Rodar os testes

```bash
go test ./...
```
