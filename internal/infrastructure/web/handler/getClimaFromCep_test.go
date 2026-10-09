package handler

import (
	"climaCepOtel/internal/usecase"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestHandler_Get(t *testing.T) {
	success := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Host {
		case "viacep.com.br":
			return newResponse(http.StatusOK, `{"cep":"12345-678","localidade":"Sao Paulo"}`), nil
		case "api.weatherapi.com":
			return newResponse(http.StatusOK, `{"current":{"temp_c":28.5}}`), nil
		default:
			return nil, errors.New("unexpected host: " + req.URL.Host)
		}
	})

	tests := []struct {
		name       string
		query      string
		transport  roundTripFunc
		wantStatus int
		wantBody   *usecase.WeatherOutputDTO
	}{
		{
			name:       "success returns weather json",
			query:      "cep=12345678",
			transport:  success,
			wantStatus: http.StatusOK,
			wantBody: &usecase.WeatherOutputDTO{
				City:           "Sao Paulo",
				TempCelsius:    28.5,
				TempFahrenheit: 83.3,
				TempKelvin:     301.5,
			},
		},
		{
			name:  "cep not found returns Not Found",
			query: "cep=99999999",
			transport: func(req *http.Request) (*http.Response, error) {
				switch req.URL.Host {
				case "viacep.com.br":
					// resposta real do ViaCEP para CEP inexistente (HTTP 200)
					return newResponse(http.StatusOK, `{"erro":"true"}`), nil
				default:
					// weather nao deve ser chamado quando o cep nao existe
					return nil, errors.New("weather should not be called for not found cep")
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:  "invalid cep returns Unprocessable Entity",
			query: "cep=invalid",
			transport: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("should not be called")
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:  "missing cep returns Unprocessable Entity",
			query: "",
			transport: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("should not be called")
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{Transport: tt.transport}
			tracer := otel.Tracer("test")
			getClima := usecase.NewUseCaseClimaLocation(client, tracer, "fake-key")
			getCep := usecase.NewUseCaseLocationCep(client, tracer)
			h := NewClimaHandler(getClima, getCep, tracer)
			req := httptest.NewRequest(http.MethodGet, "/cep?"+tt.query, nil)
			rec := httptest.NewRecorder()

			h.Get(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantBody == nil {
				return
			}

			assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
			var got usecase.WeatherOutputDTO
			require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
			assert.Equal(t, tt.wantBody.City, got.City)
			assert.InDelta(t, tt.wantBody.TempCelsius, got.TempCelsius, 0.001)
			assert.InDelta(t, tt.wantBody.TempFahrenheit, got.TempFahrenheit, 0.001)
			assert.InDelta(t, tt.wantBody.TempKelvin, got.TempKelvin, 0.001)
		})
	}
}
