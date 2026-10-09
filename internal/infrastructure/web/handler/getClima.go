package handler

import (
	"climaCepOtel/internal/usecase"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
)

const DEADLINE = time.Second * 5

type ClimaHandler struct {
	climaLocation *usecase.UsecaseClimaLocation
	locationCep   *usecase.LocationFromCep
	tracer        trace.Tracer
}

func NewClimaHandler(climaLocation *usecase.UsecaseClimaLocation, locationCep *usecase.LocationFromCep, tracer trace.Tracer) *ClimaHandler {
	return &ClimaHandler{
		climaLocation: climaLocation,
		locationCep:   locationCep,
		tracer:        tracer,
	}
}

func (h *ClimaHandler) Get(w http.ResponseWriter, r *http.Request) {
	var dto usecase.CepInputDTO
	dto.Cep = r.URL.Query().Get("cep")

	ctx, cancel := context.WithTimeout(r.Context(), DEADLINE)
	defer cancel()

	ctx, span := h.tracer.Start(ctx, "Get Clima Location")
	defer span.End()

	cep, useCEPErr := h.locationCep.Execute(ctx, dto)
	if useCEPErr != nil {
		errorHandler(toAppError(useCEPErr), w, span)
		return
	}

	weather, useErr := h.climaLocation.Execute(ctx, cep.City)
	if useErr != nil {
		errorHandler(toAppError(useErr), w, span)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(weather)
	if err != nil {
		errorHandler(toAppError(err), w, span)
		return
	}
}
