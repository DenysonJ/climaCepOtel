package handler

import (
	"climaCepOtel/internal/usecase"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const DEADLINE = time.Second * 5

type ClimaHandler struct {
	climaLocation *usecase.UsecaseClimaLocation
	locationCep   *usecase.LocationFromCep
}

func NewClimaHandler(climaLocation *usecase.UsecaseClimaLocation, locationCep *usecase.LocationFromCep) *ClimaHandler {
	return &ClimaHandler{
		climaLocation: climaLocation,
		locationCep:   locationCep,
	}
}

func (h *ClimaHandler) Get(w http.ResponseWriter, r *http.Request) {
	var dto usecase.CepInputDTO
	dto.Cep = r.URL.Query().Get("cep")

	ctx, cancel := context.WithTimeout(r.Context(), DEADLINE)
	defer cancel()

	cep, useCEPErr := h.locationCep.Execute(ctx, dto)
	if useCEPErr != nil {
		errorHandler(useCEPErr, w)
		return
	}

	weather, useErr := h.climaLocation.Execute(ctx, cep.City)
	if useErr != nil {
		errorHandler(useErr, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(weather)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
