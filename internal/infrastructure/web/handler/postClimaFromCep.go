package handler

import (
	cep "climaCepOtel/internal/domain/vo"
	"climaCepOtel/internal/usecase"
	"climaCepOtel/pkgs/httputils"
	"context"
	"encoding/json"
	"errors"
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
		errorHandler(readErr, w)
		return
	}

	cepVO, cepErr := cep.New(dto.Cep)
	if cepErr != nil {
		errorHandler(cepErr, w)
		return
	}

	weather, weatherErr := h.httpClient.GetJson(ctx, fmt.Sprintf("%s/clima?cep=%s", h.externalCallURL, cepVO.Value))
	if weatherErr != nil {
		errorHandler(weatherErr, w)
		return
	}
	// weather ja e um JSON ([]byte) vindo do microsservico clima;
	// repassa cru para nao re-encodar (json.Encode de []byte viraria base64).
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, writeErr := w.Write(weather); writeErr != nil {
		log.Printf("writing response: %v", writeErr)
	}
}

func errorHandler(err error, w http.ResponseWriter) {
	if errors.Is(err, context.DeadlineExceeded) {
		http.Error(w, err.Error(), http.StatusRequestTimeout)
		return
	}
	if errors.Is(err, cep.ErrorInvalidCep) {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}
