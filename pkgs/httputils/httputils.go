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

func (h *HttpUtils) GetJson(ctx context.Context, url string) ([]byte, int, error) {
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if reqErr != nil {
		return nil, 0, reqErr
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, doErr := h.httpClient.Do(req)
	if doErr != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("timeout calling server (300ms exceeded): %w", doErr)
			return nil, 0, context.DeadlineExceeded
		}
		log.Println("calling server: %w", doErr)
		return nil, 0, doErr
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Println("reading response body: %w", readErr)
		return nil, resp.StatusCode, readErr
	}

	return body, resp.StatusCode, nil
}
