package main

import (
	"climaCepOtel/configs"
	"climaCepOtel/internal/infrastructure/web/handler"
	"climaCepOtel/pkgs/httputils"
	"climaCepOtel/pkgs/telemetry"
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	tracer := telemetry.Tracer(config.OtelServiceName)

	httpUtils := httputils.NewHttpUtils(telemetry.NewHTTPClient())
	cepHandler := handler.NewLocationHandler(httpUtils, tracer, config.ExternalCallURL)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/cep", cepHandler.Post)
	r.Handle("/metrics", promhttp.Handler())

	serverErr := http.ListenAndServe(":"+config.WebServerPort, telemetry.WrapHandler(r, config.OtelServiceName))
	if serverErr != nil {
		log.Fatal(serverErr)
	}
}
