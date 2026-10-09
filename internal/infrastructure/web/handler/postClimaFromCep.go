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
)

type LocationHandler struct {
	httpClient      *httputils.HttpUtils
	externalCallURL string
}

func NewLocationHandler(httpClient *httputils.HttpUtils, externalCallURL string) *LocationHandler {
	return &LocationHandler{
		httpClient:      httpClient,
		externalCallURL: externalCallURL,
	}
}

func (h *LocationHandler) Post(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DEADLINE)
	defer cancel()

	var dto usecase.CepInputDTO
	readErr := json.NewDecoder(r.Body).Decode(&dto)
	if readErr != nil {
		errorHandler(toAppError(readErr), w)
		return
	}

	cepVO, cepErr := cep.New(dto.Cep)
	if cepErr != nil {
		errorHandler(toAppError(cepErr), w)
		return
	}

	weather, status, weatherErr := h.httpClient.GetJson(ctx, fmt.Sprintf("%s/clima?cep=%s", h.externalCallURL, cepVO.Value))
	if weatherErr != nil {
		errorHandler(toAppError(weatherErr), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, writeErr := w.Write(weather); writeErr != nil {
		log.Printf("writing response: %v", writeErr)
	}
}
