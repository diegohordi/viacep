package viacep_test

import (
	"context"
	"fmt"
	"github.com/diegohordi/viacep"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func MustLoadTestDataFile(t *testing.T, fileName string) []byte {
	t.Helper()
	content, err := os.ReadFile(fmt.Sprintf("./test/testdata/%s", fileName))
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func Test_Consultar(t *testing.T) {
	type fields struct {
		baseURL    string
		httpClient func() *http.Client
	}
	type args struct {
		ctx func() (context.Context, context.CancelFunc)
		cep string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    viacep.CEP
		wantErr bool
	}{
		{
			name: "should return a valid CEP",
			fields: fields{
				baseURL: "http://viacep.com.br/ws/",
				httpClient: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(MustLoadTestDataFile(t, "valid_result.json"))
							return resp.Result()
						}),
						Timeout: 3 * time.Second,
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				cep: "01001000",
			},
			want: viacep.CEP{
				CEP:         "01001-000",
				Logradouro:  "Praça da Sé",
				Complemento: "lado ímpar",
				Bairro:      "Sé",
				Localidade:  "São Paulo",
				UF:          "SP",
				IBGE:        "3550308",
				GIA:         "1004",
				DDD:         "11",
				SIAFI:       "7107",
				Erro:        false,
			},
			wantErr: false,
		},
		{
			name: "should return an error due to the invalid CEP",
			fields: fields{
				baseURL: "http://viacep.com.br/ws/",
				httpClient: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.WriteString("")
							resp.WriteHeader(http.StatusBadRequest)
							return resp.Result()
						}),
						Timeout: 3 * time.Second,
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				cep: "XXXXXX",
			},
			want:    viacep.CEP{},
			wantErr: true,
		},
		{
			name: "should return an error due to a not found CEP",
			fields: fields{
				baseURL: "http://viacep.com.br/ws/",
				httpClient: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(MustLoadTestDataFile(t, "not_found_result.json"))
							return resp.Result()
						}),
						Timeout: 3 * time.Second,
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				cep: "99999999",
			},
			want:    viacep.CEP{},
			wantErr: true,
		},
		{
			name: "should return an error due to an unknown response body",
			fields: fields{
				baseURL: "http://viacep.com.br/ws/",
				httpClient: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.WriteString("[]")
							return resp.Result()
						}),
						Timeout: 3 * time.Second,
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				cep: "01001000",
			},
			want:    viacep.CEP{},
			wantErr: true,
		},
		{
			name: "should return an error due to context timeout",
			fields: fields{
				baseURL: "http://viacep.com.br/ws/",
				httpClient: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							time.Sleep(10 * time.Second)
							return &http.Response{}
						}),
						Timeout: 3 * time.Second,
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.WithTimeout(context.TODO(), 1*time.Millisecond)
				},
				cep: "01001000",
			},
			want:    viacep.CEP{},
			wantErr: true,
		},
		{
			name: "should return an error due to client timeout",
			fields: fields{
				baseURL: "http://viacep.com.br/ws/",
				httpClient: func() *http.Client {
					return &http.Client{
						Timeout: 1 * time.Millisecond,
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				cep: "01001000",
			},
			want:    viacep.CEP{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := viacep.NewClient(tt.fields.baseURL, tt.fields.httpClient())
			ctx, cancel := tt.args.ctx()
			if cancel != nil {
				defer cancel()
			}
			got, err := d.Consultar(ctx, tt.args.cep)
			if (err != nil) != tt.wantErr {
				t.Errorf("Consultar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Consultar() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Integration_Consultar(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests...")
	}
	type fields struct {
		baseURL    string
		httpClient func() *http.Client
	}
	type args struct {
		ctx func() (context.Context, context.CancelFunc)
		cep string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    viacep.CEP
		wantErr bool
	}{
		{
			name: "should return a valid CEP",
			fields: fields{
				baseURL: "http://viacep.com.br/ws/",
				httpClient: func() *http.Client {
					return &http.Client{Timeout: 3 * time.Second}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				cep: "01001000",
			},
			want: viacep.CEP{
				CEP:         "01001-000",
				Logradouro:  "Praça da Sé",
				Complemento: "lado ímpar",
				Bairro:      "Sé",
				Localidade:  "São Paulo",
				UF:          "SP",
				IBGE:        "3550308",
				GIA:         "1004",
				DDD:         "11",
				SIAFI:       "7107",
				Erro:        false,
			},
			wantErr: false,
		},
		{
			name: "should return an error due to the invalid CEP",
			fields: fields{
				baseURL: "http://viacep.com.br/ws/",
				httpClient: func() *http.Client {
					return &http.Client{Timeout: 3 * time.Second}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				cep: "XXXXXX",
			},
			want:    viacep.CEP{},
			wantErr: true,
		},
		{
			name: "should return an error due to a not found CEP",
			fields: fields{
				baseURL: "http://viacep.com.br/ws/",
				httpClient: func() *http.Client {
					return &http.Client{Timeout: 3 * time.Second}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				cep: "99999999",
			},
			want:    viacep.CEP{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := viacep.NewClient(tt.fields.baseURL, tt.fields.httpClient())
			ctx, cancel := tt.args.ctx()
			if cancel != nil {
				defer cancel()
			}
			got, err := d.Consultar(ctx, tt.args.cep)
			if (err != nil) != tt.wantErr {
				t.Errorf("Consultar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Consultar() got = %v, want %v", got, tt.want)
			}
		})
	}
}
