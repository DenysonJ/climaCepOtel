package handler

import (
	cep "climaCepOtel/internal/domain/vo"
	"climaCepOtel/internal/usecase"
	"climaCepOtel/pkgs/httputils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

type LocationHandler struct {
	httpClient      *httputils.HttpUtils
	tracer          trace.Tracer
	externalCallURL string
}

func NewLocationHandler(httpClient *httputils.HttpUtils, tracer trace.Tracer, externalCallURL string) *LocationHandler {
	return &LocationHandler{
		httpClient:      httpClient,
		tracer:          tracer,
		externalCallURL: externalCallURL,
	}
}

func (h *LocationHandler) Post(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DEADLINE)
	defer cancel()

	ctx, span := h.tracer.Start(ctx, "Post Clima Location")
	defer span.End()

	var dto usecase.CepInputDTO
	readErr := json.NewDecoder(r.Body).Decode(&dto)
	if readErr != nil {
		errorHandler(toAppError(readErr), w, span)
		return
	}

	cepVO, cepErr := cep.New(dto.Cep)
	if cepErr != nil {
		errorHandler(toAppError(cepErr), w, span)
		return
	}

	weather, status, weatherErr := h.httpClient.GetJson(ctx, fmt.Sprintf("%s/clima?cep=%s", h.externalCallURL, cepVO.Value))
	if weatherErr != nil {
		errorHandler(toAppError(weatherErr), w, span)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, writeErr := w.Write(weather); writeErr != nil {
		log.Printf("writing response: %v", writeErr)
		errorHandler(toAppError(writeErr), w, span)
	}
}
