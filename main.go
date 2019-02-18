package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
)

const (
	digipixEndpoint = "https://service-homolog.digipix.com.br/v0b/shipments/zipcode/"
	jwtToken        = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJkZXNmaW8uZm90b3JlZ2lzdHJvLmNvbS5iciIsImV4cCI6MTU3NzU1NDEzMywianRpIjoiNzBlODRlZmQtMGRmNC00ZmZhLTlmYTYtNTI1M2ZjNmFmMDgyIiwiaWF0IjoxNTQ2NDUwMTMzLCJpc3MiOiJodHRwczovL3NlcnZpY2UtaG9tb2xvZy5kaWdpcGl4LmNvbS5iciIsInN0b3JlSWQiOjc5LCJzdG9yZU5hbWUiOiJGb3RvcmVnaXN0cm8iLCJzdG9yZVVSTCI6ImRlc2Zpby5mb3RvcmVnaXN0cm8uY29tLmJyIn0.yPFKdRdc4jTAUuziqfkvJm74W5axDelkaH-Q6lBTE8k"
)

func main() {
	http.HandleFunc("/consultar", ConsultaCEP)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//ConsultaCep recebe o cep e realiza a consulta no endpoint da Digipix
func ConsultaCEP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Inicialdo consulta de CEP...")

	//Valida e trata o cep informado
	cepField := r.FormValue("cep")
	numbers, err := regexp.Compile("[^0-9]+")
	if err != nil {
		ServerError(w, fmt.Sprintf("Erro ao compilar a regex: %v", err))
		return
	}
	cep := numbers.ReplaceAllString(cepField, "")
	if cep == "" {
		http.Error(w, "CEP não preenchido", http.StatusBadRequest)
		return
	}
	if len(cep) != 8 {
		http.Error(w, "Tamanho inválido do CEP", http.StatusBadRequest)
		return
	}
	log.Printf("Buscando CEP '%s'...", cep)

	//Monta o endereço de consulta
	url := fmt.Sprintf("%s%s", digipixEndpoint, cep)
	log.Printf("Enviando requisição para: %s", url)

	//Constrói a requisição
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		ServerError(w, fmt.Sprintf("Erro ao construir a requisição: %v", err))
		return
	}

	//Adiciona o token de autenticação JWT
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	//Realiza a requisição de consulta do CEP
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ServerError(w, fmt.Sprintf("Erro ao enviar requisição ao endpoint digipix: %v", err))
		return
	}

	//Faz o dump do response
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		ServerError(w, fmt.Sprintf("Erro ao realizar o dump do response: %v", err))
		return
	}
	log.Printf("Response: %s", dump)

	//Analisa o status de retorno
	switch resp.StatusCode {
	case http.StatusOK:
		//Decodifica o endereço recebido
		var endereco Address
		err = json.NewDecoder(resp.Body).Decode(&endereco)
		if err != nil {
			ServerError(w, fmt.Sprintf("Erro ao decodificar o json de retorno: %v", err))
			return
		}
		log.Printf("Endereco: %#v", endereco)

		//Retorna o json do endereço recebido
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(w).Encode(endereco)
		if err != nil {
			ServerError(w, fmt.Sprintf("Erro ao escrever o json no response: %v", err))
			return
		}
	case http.StatusUnauthorized:
		w.Write([]byte("Acesso não autorizado"))
		return
	case http.StatusNotFound:
		w.Write([]byte("CEP não encontrado"))
		return
	default:
		ServerError(w, fmt.Sprintf("Status de retorno não mapeado: %v", resp.StatusCode))
		return
	}
	log.Printf("Consulta de CEP concluída")
}

//ServerError retorna status InternalServerError e loga o erro
func ServerError(w http.ResponseWriter, msg string) {
	http.Error(w, "Erro interno", http.StatusInternalServerError)
	log.Printf(msg)
}

//Address estrutura de retorno da consulta de CEP
type Address struct {
	State          string `json:"state"`
	City           string `json:"city"`
	Neighborhood   string `json:"neighborhood"`
	Street         string `json:"street"`
	IBGE           string `json:"ibge"`
	AdditionalInfo string `json:"additional_info"`
	Bairro         string `json:"bairro"`
}
