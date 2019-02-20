package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type CenarioDeTeste struct {
	Cep             string
	StatusDeRetorno int
	Endereco        Address
	TemErro         bool
}

func TestConsultarEndereco(t *testing.T) {
	//Cria o servidor de teste (mock para o recurso de consulta da Digipix)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var cep string
		partes := strings.Split(r.RequestURI, "/")
		if len(partes) > 1 {
			cep = partes[1]
		}
		log.Printf(cep)
		switch cep {
		case "":
			http.Error(w, "Bad Request", http.StatusBadRequest)
		case "12234576890963":
			http.Error(w, "Bad Request", http.StatusBadRequest)
		case "04147020":
			end := Address{
				State:        "SP",
				City:         "São Paulo",
				Neighborhood: "Vila do Bosque",
				Street:       "R Alfredo de S O Netto",
			}
			_ = json.NewEncoder(w).Encode(&end)
		case "00000000":
			http.Error(w, "Não encontrado", http.StatusNotFound)
		case "06381340":
			http.Error(w, "Não autorizado", http.StatusUnauthorized)
		default:
			http.Error(w, "Não implementado", http.StatusNotImplemented)
		}
	}))
	defer ts.Close()

	//Sobrescreve o serviço com a URL do mock e volta com o valor original no final
	original := DigipixURL
	DigipixURL = fmt.Sprintf("%s/", ts.URL)
	defer func() {
		DigipixURL = original
	}()

	//Monta os cenários de teste
	cenarios := []CenarioDeTeste{
		//Cenário 0: cep vazio
		CenarioDeTeste{
			Cep:             " ",
			StatusDeRetorno: 400,
			Endereco:        Address{},
			TemErro:         true,
		},
		//Cenário 1: cep inválido
		CenarioDeTeste{
			Cep:             "122345fadsnf76890963sdf",
			StatusDeRetorno: 400,
			Endereco:        Address{},
			TemErro:         true,
		},
		//Cenário 2: endereco encontrado
		CenarioDeTeste{
			Cep:             "04147020",
			StatusDeRetorno: 200,
			Endereco: Address{
				State:        "SP",
				City:         "São Paulo",
				Neighborhood: "Vila do Bosque",
				Street:       "R Alfredo de S O Netto",
			},
			TemErro: false,
		},
		//Cenário 3: endereço não encontrado
		CenarioDeTeste{
			Cep:             "00000000",
			StatusDeRetorno: 404,
			Endereco:        Address{},
			TemErro:         true,
		},
		//Cenário 4: acesso não autorizado
		CenarioDeTeste{
			Cep:             "06381340",
			StatusDeRetorno: 401,
			Endereco:        Address{},
			TemErro:         true,
		},
	}
	for i, c := range cenarios {
		t.Logf("Cenário de teste %d", i)
		endereco, status, err := ConsultarEndereco(c.Cep)
		if c.TemErro && err == nil {
			t.Errorf("Esperado erro.")
		}
		if !c.TemErro && err != nil {
			t.Errorf("Erro inesperado: %v", err)
		}
		if status != c.StatusDeRetorno {
			t.Errorf("Status inesperado '%d', esperado '%d'", status, c.StatusDeRetorno)
		}
		if endereco.AdditionalInfo != c.Endereco.AdditionalInfo {
			t.Errorf("AdditionalInfo inesperado '%s', esperado '%s'", endereco.AdditionalInfo, c.Endereco.AdditionalInfo)
		}
		if endereco.Bairro != c.Endereco.Bairro {
			t.Errorf("Bairro inesperado '%s', esperado '%s'", endereco.Bairro, c.Endereco.Bairro)
		}
		if endereco.City != c.Endereco.City {
			t.Errorf("City inesperado '%s', esperado '%s'", endereco.City, c.Endereco.City)
		}
		if endereco.IBGE != c.Endereco.IBGE {
			t.Errorf("IBGE inesperado '%s', esperado '%s'", endereco.IBGE, c.Endereco.IBGE)
		}
		if endereco.Neighborhood != c.Endereco.Neighborhood {
			t.Errorf("Neighborhood inesperado '%s', esperado '%s'", endereco.Neighborhood, c.Endereco.Neighborhood)
		}
		if endereco.State != c.Endereco.State {
			t.Errorf("State inesperado '%s', esperado '%s'", endereco.State, c.Endereco.State)
		}
		if endereco.Street != c.Endereco.Street {
			t.Errorf("Street inesperado '%s', esperado '%s'", endereco.Street, c.Endereco.Street)
		}
	}
}
