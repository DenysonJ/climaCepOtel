package httputils

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
)

type HttpUtils struct {
	httpClient *http.Client
}

func NewHttpUtils(httpClient *http.Client) *HttpUtils {
	return &HttpUtils{
		httpClient: httpClient,
	}
}

func (h *HttpUtils) GetJson(ctx context.Context, url string) ([]byte, error) {
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if reqErr != nil {
		return nil, reqErr
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, doWeatherErr := h.httpClient.Do(req)
	if doWeatherErr != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("timeout calling server (300ms exceeded): %w", doWeatherErr)
			return nil, context.DeadlineExceeded
		}
		log.Println("calling server: %w", doWeatherErr)
		return nil, doWeatherErr
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Println("reading response body: %w", readErr)
		return nil, readErr
	}

	return body, nil
}
