package usecase

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

func TestUsecaseClimaLocation_Execute(t *testing.T) {
	tests := []struct {
		name      string
		city      string
		transport roundTripFunc
		want      WeatherOutputDTO
		wantErr   bool
	}{
		{
			name: "success returns converted temperatures",
			city: "Sao Paulo",
			transport: func(req *http.Request) (*http.Response, error) {
				return newResponse(http.StatusOK, `{"current":{"temp_c":28.5}}`), nil
			},
			want: WeatherOutputDTO{
				City:           "Sao Paulo",
				TempCelsius:    28.5,
				TempFahrenheit: 83.3,
				TempKelvin:     301.5,
			},
			wantErr: false,
		},
		{
			name: "success with negative temperature",
			city: "Curitiba",
			transport: func(req *http.Request) (*http.Response, error) {
				return newResponse(http.StatusOK, `{"current":{"temp_c":-5}}`), nil
			},
			want: WeatherOutputDTO{
				City:           "Curitiba",
				TempCelsius:    -5,
				TempFahrenheit: 23,
				TempKelvin:     268,
			},
			wantErr: false,
		},
		{
			name: "weather request error",
			city: "Sao Paulo",
			transport: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("weather down")
			},
			wantErr: true,
		},
		{
			name: "weather invalid json body",
			city: "Sao Paulo",
			transport: func(req *http.Request) (*http.Response, error) {
				return newResponse(http.StatusOK, `not a json`), nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUseCaseClimaLocation(&http.Client{Transport: tt.transport}, otel.Tracer("test"), "fake-key")

			got, err := uc.Execute(context.Background(), tt.city)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, WeatherOutputDTO{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.City, got.City)
			assert.InDelta(t, tt.want.TempCelsius, got.TempCelsius, 0.001)
			assert.InDelta(t, tt.want.TempFahrenheit, got.TempFahrenheit, 0.001)
			assert.InDelta(t, tt.want.TempKelvin, got.TempKelvin, 0.001)
		})
	}
}
