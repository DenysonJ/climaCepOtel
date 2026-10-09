package usecase

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestLocationFromCep_Execute(t *testing.T) {
	cepOK := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return newResponse(http.StatusOK, `{"cep":"12345-678","localidade":"Sao Paulo"}`), nil
	})

	tests := []struct {
		name      string
		cep       string
		transport roundTripFunc
		want      CepOutputDTO
		wantErr   bool
		errIs     error
	}{
		{
			name:      "success returns cep and city",
			cep:       "12345678",
			transport: cepOK,
			want: CepOutputDTO{
				Cep:  "12345-678",
				City: "Sao Paulo",
			},
			wantErr: false,
		},
		{
			name: "cep not found returns ErrorCepNotFound",
			cep:  "99999999",
			transport: func(req *http.Request) (*http.Response, error) {
				// resposta real do ViaCEP para CEP inexistente (HTTP 200)
				return newResponse(http.StatusOK, `{"erro":"true"}`), nil
			},
			wantErr: true,
			errIs:   ErrorCepNotFound,
		},
		{
			name: "invalid cep does not call viacep",
			cep:  "invalid",
			transport: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("should not be called")
			},
			wantErr: true,
		},
		{
			name: "viacep request error",
			cep:  "12345678",
			transport: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("viacep down")
			},
			wantErr: true,
		},
		{
			name: "viacep invalid json body",
			cep:  "12345678",
			transport: func(req *http.Request) (*http.Response, error) {
				return newResponse(http.StatusOK, `not a json`), nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUseCaseLocationCep(&http.Client{Transport: tt.transport})

			got, err := uc.Execute(context.Background(), CepInputDTO{Cep: tt.cep})

			if tt.wantErr {
				require.Error(t, err)
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs)
				}
				assert.Equal(t, CepOutputDTO{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertCelsiusToFahrenheit(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
		want    float64
	}{
		{name: "zero celsius", celsius: 0, want: 32},
		{name: "positive celsius", celsius: 28.5, want: 83.3},
		{name: "negative celsius", celsius: -40, want: -40},
		{name: "boiling point", celsius: 100, want: 212},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertCelsiusToFahrenheit(tt.celsius)

			assert.InDelta(t, tt.want, got, 0.0001)
		})
	}
}

func TestConvertCelsiusToKelvin(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
		want    float64
	}{
		{name: "zero celsius", celsius: 0, want: 273},
		{name: "positive celsius", celsius: 28.5, want: 301.5},
		{name: "negative celsius", celsius: -273, want: 0},
		{name: "boiling point", celsius: 100, want: 373},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertCelsiusToKelvin(tt.celsius)

			assert.InDelta(t, tt.want, got, 0.0001)
		})
	}
}
