package usecase

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				TempCelsius:    "28.50",
				TempFahrenheit: "83.30",
				TempKelvin:     "301.50",
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
				TempCelsius:    "-5.00",
				TempFahrenheit: "23.00",
				TempKelvin:     "268.00",
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
			uc := NewUseCaseClimaLocation(&http.Client{Transport: tt.transport}, "fake-key")

			got, err := uc.Execute(context.Background(), tt.city)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, WeatherOutputDTO{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
