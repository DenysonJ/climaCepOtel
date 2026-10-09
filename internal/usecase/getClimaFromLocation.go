package usecase

import (
	"climaCepOtel/pkgs/httputils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	neturl "net/url"

	"go.opentelemetry.io/otel/trace"
)

const WEATHERAPI = "http://api.weatherapi.com/v1/current.json"

type CityInputDTO struct {
	City string `json:"city"`
}

type WeatherOutputDTO struct {
	City           string  `json:"city"`
	TempCelsius    float64 `json:"temp_C"`
	TempFahrenheit float64 `json:"temp_F"`
	TempKelvin     float64 `json:"temp_K"`
}

type WeatherAPIResponse struct {
	Current struct {
		LastUpdated string  `json:"last_updated"`
		TempC       float64 `json:"temp_c"`
	} `json:"current"`
}

type UsecaseClimaLocation struct {
	httpClient    *httputils.HttpUtils
	tracer        trace.Tracer
	weatherApiKey string
}

func NewUseCaseClimaLocation(httpClient *http.Client, tracer trace.Tracer, apiKey string) *UsecaseClimaLocation {
	return &UsecaseClimaLocation{
		httpClient:    httputils.NewHttpUtils(httpClient),
		tracer:        tracer,
		weatherApiKey: apiKey,
	}
}

func (u *UsecaseClimaLocation) Execute(ctx context.Context, city string) (WeatherOutputDTO, error) {
	ctx, span := u.tracer.Start(ctx, "Execute Get Clima Location")
	defer span.End()

	url := WEATHERAPI + "?key=" + u.weatherApiKey + "&q=" + neturl.QueryEscape(city)
	respWeather, _, doWeatherErr := u.httpClient.GetJson(ctx, url)
	if doWeatherErr != nil {
		return WeatherOutputDTO{}, doWeatherErr
	}

	var weatherResp WeatherAPIResponse
	if unmarshalErr := json.Unmarshal(respWeather, &weatherResp); unmarshalErr != nil {
		log.Println("unmarshalling response body Weather: %w", unmarshalErr)
		return WeatherOutputDTO{}, unmarshalErr
	}

	currentCelsius := weatherResp.Current.TempC

	return WeatherOutputDTO{
		City:           city,
		TempCelsius:    currentCelsius,
		TempFahrenheit: convertCelsiusToFahrenheit(currentCelsius),
		TempKelvin:     convertCelsiusToKelvin(currentCelsius),
	}, nil
}

func convertCelsiusToFahrenheit(celsius float64) float64 {
	return (celsius * 1.8) + 32.0
}

func convertCelsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}
