// Package viacep contains a httpClient used to access the ViaCEP services.
package viacep

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Error represents the errors returned by this client.
type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrCEPInvalid  = Error("the given CEP is invalid")
	ErrCEPNotFound = Error("the given CEP was not found")
)

// CEP holds information about its address.
type CEP struct {
	CEP         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	UF          string `json:"uf"`
	IBGE        string `json:"ibge"`
	GIA         string `json:"gia"`
	DDD         string `json:"ddd"`
	SIAFI       string `json:"siafi"`
	Erro        bool   `json:"erro"`
}

// Client defines the ViaCEP services available.
type Client interface {

	// Consultar searches the given CEP and return its address or an error if something goes wrong during the
	// request, or even if the given CEP is invalid or it was not found.
	Consultar(ctx context.Context, cep string) (CEP, error)
}

type defaultClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new httpClient for the ViaCEP services.
func NewClient(baseURL string, httpClient *http.Client) Client {
	return &defaultClient{baseURL: baseURL, httpClient: httpClient}
}

func (d defaultClient) parseResponse(resp *http.Response) (CEP, error) {
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode == http.StatusBadRequest {
		return CEP{}, ErrCEPInvalid
	}
	var cep CEP
	if err := json.NewDecoder(resp.Body).Decode(&cep); err != nil {
		return CEP{}, err
	}
	if cep.Erro {
		return CEP{}, ErrCEPNotFound
	}
	return cep, nil
}

func (d defaultClient) Consultar(ctx context.Context, cep string) (CEP, error) {
	errChan := make(chan error, 1)
	resultChan := make(chan CEP, 1)
	go func() {
		var resp *http.Response
		serviceURL := fmt.Sprintf("%s/%s/json", d.baseURL, cep)
		resp, err := d.httpClient.Get(serviceURL)
		if err != nil {
			errChan <- err
			return
		}
		result, err := d.parseResponse(resp)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()
	select {
	case err := <-errChan:
		return CEP{}, err
	case <-ctx.Done():
		return CEP{}, ctx.Err()
	case result := <-resultChan:
		return result, nil
	}
}
