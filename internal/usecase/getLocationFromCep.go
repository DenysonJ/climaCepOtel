package usecase

import (
	cep "climaCepOtel/internal/domain/vo"
	"climaCepOtel/pkgs/httputils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const VIACEPURL = "http://viacep.com.br/ws/%s/json"

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

	respCEP, doCEPErr := u.httpClient.GetJson(ctx, fmt.Sprintf(VIACEPURL, cepVO.Value))
	if doCEPErr != nil {
		return CepOutputDTO{}, doCEPErr
	}

	var cepJson CepResponse
	if unmarshalErr := json.Unmarshal(respCEP, &cepJson); unmarshalErr != nil {
		log.Println("unmarshalling response body CEP: %w", unmarshalErr)
		return CepOutputDTO{}, unmarshalErr
	}

	return CepOutputDTO{
		Cep:  cepJson.Cep,
		City: cepJson.Localidade,
	}, nil
}
