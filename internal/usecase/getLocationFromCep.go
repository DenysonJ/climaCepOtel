package usecase

import (
	cep "climaCepOtel/internal/domain/vo"
	"climaCepOtel/pkgs/httputils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

const VIACEPURL = "http://viacep.com.br/ws/%s/json"

var ErrorCepNotFound = errors.New("can not find zipcode")

type CepInputDTO struct {
	Cep string
}

type CepOutputDTO struct {
	Cep  string `json:"cep"`
	City string `json:"city"`
}

type CepResponse struct {
	Cep        string `json:"cep"`
	Localidade string `json:"localidade"`
	Error      string `json:"erro"`
}

type LocationFromCep struct {
	httpClient *httputils.HttpUtils
	tracer     trace.Tracer
}

func NewUseCaseLocationCep(httpClient *http.Client, tracer trace.Tracer) *LocationFromCep {
	return &LocationFromCep{
		httpClient: httputils.NewHttpUtils(httpClient),
		tracer:     tracer,
	}
}

func (u *LocationFromCep) Execute(ctx context.Context, dto CepInputDTO) (CepOutputDTO, error) {
	ctx, span := u.tracer.Start(ctx, "Execute LocationFromCep")
	defer span.End()

	cepVO, cepErr := cep.New(dto.Cep)
	if cepErr != nil {
		return CepOutputDTO{}, cepErr
	}

	respCEP, _, doCEPErr := u.httpClient.GetJson(ctx, fmt.Sprintf(VIACEPURL, cepVO.Value))
	if doCEPErr != nil {
		return CepOutputDTO{}, doCEPErr
	}

	var cepJson CepResponse
	if unmarshalErr := json.Unmarshal(respCEP, &cepJson); unmarshalErr != nil {
		log.Println("unmarshalling response body CEP: %w", unmarshalErr)
		return CepOutputDTO{}, unmarshalErr
	}

	if cepJson.Error != "" {
		return CepOutputDTO{}, ErrorCepNotFound
	}

	return CepOutputDTO{
		Cep:  cepJson.Cep,
		City: cepJson.Localidade,
	}, nil
}
