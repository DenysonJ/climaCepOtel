package main

import (
	"climaCepOtel/configs"
	"climaCepOtel/internal/infrastructure/web/handler"
	"climaCepOtel/internal/usecase"
	"climaCepOtel/pkgs/telemetry"
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	config, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	shutdown, otelErr := telemetry.InitTracer(context.Background(), telemetry.Config{
		ServiceName:       config.OtelServiceName,
		CollectorEndpoint: config.OtelExporterEndpoint,
	})
	if otelErr != nil {
		log.Fatal(otelErr)
	}
	defer func() {
		if shutdownErr := shutdown(context.Background()); shutdownErr != nil {
			log.Printf("shutting down tracer: %v", shutdownErr)
		}
	}()

	httpClient := telemetry.NewHTTPClient()
	getClima := usecase.NewUseCaseClimaLocation(httpClient, config.WeatherApiKey)
	getCep := usecase.NewUseCaseLocationCep(httpClient)
	climaHandler := handler.NewClimaHandler(getClima, getCep)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/clima", climaHandler.Get)
	serverErr := http.ListenAndServe(":"+config.WebServerPort, telemetry.WrapHandler(r, config.OtelServiceName))
	if serverErr != nil {
		log.Fatal(serverErr)
	}
}
