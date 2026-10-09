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
}

func NewUseCaseLocationCep(httpClient *http.Client) *LocationFromCep {
	return &LocationFromCep{
		httpClient: httputils.NewHttpUtils(httpClient),
	}
}

func (u *LocationFromCep) Execute(ctx context.Context, dto CepInputDTO) (CepOutputDTO, error) {
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
