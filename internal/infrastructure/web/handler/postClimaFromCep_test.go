package handler

import (
	"climaCepOtel/pkgs/httputils"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestLocationHandler_Post(t *testing.T) {
	const externalURL = "http://clima:8181"
	const weatherJSON = `{"city":"Sao Paulo","temp_C":28.5,"temp_F":83.3,"temp_K":301.5}`

	notCalled := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("should not be called")
	})

	tests := []struct {
		name       string
		body       string
		transport  roundTripFunc
		wantStatus int
		wantBody   string
	}{
		{
			name: "valid cep returns weather json from clima service",
			body: `{"cep":"12345678"}`,
			transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return newResponse(http.StatusOK, weatherJSON), nil
			}),
			wantStatus: http.StatusOK,
			wantBody:   weatherJSON,
		},
		{
			name:       "invalid cep returns unprocessable entity",
			body:       `{"cep":"invalid"}`,
			transport:  notCalled,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "malformed body returns internal server error",
			body:       `not a json`,
			transport:  notCalled,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "clima service error returns internal server error",
			body: `{"cep":"12345678"}`,
			transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("clima down")
			}),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{Transport: tt.transport}
			h := NewLocationHandler(httputils.NewHttpUtils(client), otel.Tracer("test"), externalURL)
			req := httptest.NewRequest(http.MethodPost, "/cep", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			h.Post(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantBody == "" {
				return
			}

			assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
			assert.JSONEq(t, tt.wantBody, rec.Body.String())
		})
	}
}
