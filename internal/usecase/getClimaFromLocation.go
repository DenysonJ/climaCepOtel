package usecase

import (
	"climaCepOtel/pkgs/httputils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	neturl "net/url"
)

const WEATHERAPI = "http://api.weatherapi.com/v1/current.json"

type CityInputDTO struct {
	City string `json:"city"`
}

type WeatherOutputDTO struct {
	City           string `json:"city"`
	TempCelsius    string `json:"temp_C"`
	TempFahrenheit string `json:"temp_F"`
	TempKelvin     string `json:"temp_K"`
}

type WeatherAPIResponse struct {
	Current struct {
		LastUpdated string  `json:"last_updated"`
		TempC       float64 `json:"temp_c"`
	} `json:"current"`
}

type UsecaseClimaLocation struct {
	httpClient    *httputils.HttpUtils
	weatherApiKey string
}

func NewUseCaseClimaLocation(httpClient *http.Client, apiKey string) *UsecaseClimaLocation {
	return &UsecaseClimaLocation{
		httpClient:    httputils.NewHttpUtils(httpClient),
		weatherApiKey: apiKey,
	}
}

func (u *UsecaseClimaLocation) Execute(ctx context.Context, city string) (WeatherOutputDTO, error) {
	url := WEATHERAPI + "?key=" + u.weatherApiKey + "&q=" + neturl.QueryEscape(city)
	respWeather, doWeatherErr := u.httpClient.GetJson(ctx, url)
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
		TempCelsius:    fmt.Sprintf("%.2f", currentCelsius),
		TempFahrenheit: fmt.Sprintf("%.2f", convertCelsiusToFahrenheit(currentCelsius)),
		TempKelvin:     fmt.Sprintf("%.2f", convertCelsiusToKelvin(currentCelsius)),
	}, nil
}

func convertCelsiusToFahrenheit(celsius float64) float64 {
	return (celsius * 1.8) + 32.0
}

func convertCelsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}
