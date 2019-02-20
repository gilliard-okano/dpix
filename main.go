package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"text/template"
)

var (
	//DigipixURL url do serviço de consulta de cep
	DigipixURL = "https://service-homolog.digipix.com.br/v0b/shipments/zipcode/"
	//JwtToken token de consulta de serviço
	JwtToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJkZXNmaW8uZm90b3JlZ2lzdHJvLmNvbS5iciIsImV4cCI6MTU3NzU1NDEzMywianRpIjoiNzBlODRlZmQtMGRmNC00ZmZhLTlmYTYtNTI1M2ZjNmFmMDgyIiwiaWF0IjoxNTQ2NDUwMTMzLCJpc3MiOiJodHRwczovL3NlcnZpY2UtaG9tb2xvZy5kaWdpcGl4LmNvbS5iciIsInN0b3JlSWQiOjc5LCJzdG9yZU5hbWUiOiJGb3RvcmVnaXN0cm8iLCJzdG9yZVVSTCI6ImRlc2Zpby5mb3RvcmVnaXN0cm8uY29tLmJyIn0.yPFKdRdc4jTAUuziqfkvJm74W5axDelkaH-Q6lBTE8k"
)

func main() {
	http.HandleFunc("/", ServicoDeEndereco)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//PaginaInicial renderiza a página inicial
func PaginaInicial(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/home.html")
	if err != nil {
		log.Printf("Falha ao carregar o template: %v", err)
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}

//ServicoDeEndereco recebe o cep e realiza a consulta no endpoint da Digipix
func ServicoDeEndereco(w http.ResponseWriter, r *http.Request) {
	log.Printf("Inicialdo consulta de CEP...")

	switch r.Method {
	case "GET":
		var (
			contexto = map[string]interface{}{}
			cep      = strings.TrimSpace(r.FormValue("cep"))
			endereco Address
			status   int
			err      error
		)

		if cep != "" {
			contexto["CEP"] = cep
			endereco, status, err = ConsultarEndereco(cep)
			if err != nil {
				log.Printf("Falhou ao counsultar o endereco: %v", err)
				contexto["Msg"] = "Erro ao realizar a consulta"
			}
			if status == http.StatusOK {
				contexto["Address"] = endereco
			}
			if status == http.StatusNotFound {
				contexto["Msg"] = "Endereço não encontrado"
			}
		}
		t, err := template.ParseFiles("templates/home.html")
		if err != nil {
			log.Printf("Falha ao carregar o template: %v", err)
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}

		err = t.Execute(w, contexto)
		if err != nil {
			log.Printf("Erro ao executar o template: %v", err)
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}

		log.Printf("Consulta de CEP concluída")
	default:
		http.Error(w, "Recurso não implementado", http.StatusNotImplemented)
		return
	}
}

//ConsultarEndereco realiza a consulta do endereco no serviço da Digipix
func ConsultarEndereco(cep string) (Address, int, error) {
	var (
		endereco Address
		err      error
	)
	//Valida e trata o cep informado
	digitos, err := regexp.Compile("[^0-9]+")
	if err != nil {
		return endereco, http.StatusInternalServerError, fmt.Errorf("Erro ao compilar a regex: %v", err)
	}
	cep = digitos.ReplaceAllString(cep, "")
	if cep == "" {
		return endereco, http.StatusBadRequest, fmt.Errorf("CEP não preenchido")
	}
	if len(cep) != 8 {
		return endereco, http.StatusBadRequest, fmt.Errorf("Tamanho inválido do CEP")
	}
	log.Printf("Buscando CEP '%s'...", cep)

	//Monta o endereço de consulta
	url := fmt.Sprintf("%s%s", DigipixURL, cep)
	log.Printf("Enviando requisição para: %s", url)

	//Constrói a requisição
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return endereco, http.StatusInternalServerError, fmt.Errorf(fmt.Sprintf("Erro ao construir a requisição: %v", err))
	}

	//Adiciona o token de autenticação JWT
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", JwtToken))

	//Realiza a requisição de consulta do CEP
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return endereco, http.StatusInternalServerError, fmt.Errorf("Erro ao enviar requisição ao endpoint digipix: %v", err)
	}

	//Faz o dump do response
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return endereco, http.StatusInternalServerError, fmt.Errorf("Erro ao realizar o dump do response: %v", err)
	}
	log.Printf("Response: %s", dump)

	//Analisa o status de retorno
	switch resp.StatusCode {
	case http.StatusOK:
		//Decodifica o endereço recebido
		err = json.NewDecoder(resp.Body).Decode(&endereco)
		if err != nil {
			return endereco, http.StatusInternalServerError, fmt.Errorf("Erro ao decodificar o json de retorno: %v", err)
		}
		if endereco.NaoPreenchido() {
			return endereco, http.StatusNotFound, nil
		} else {
			log.Printf("Endereco: %#v", endereco)
		}
	case http.StatusUnauthorized:
		return endereco, http.StatusUnauthorized, fmt.Errorf("Acesso não autorizado")
	case http.StatusNotFound:
		return endereco, http.StatusNotFound, nil
	default:
		return endereco, http.StatusInternalServerError, fmt.Errorf("Status de retorno não mapeado: %v", resp.StatusCode)
	}
	return endereco, http.StatusOK, nil
}

//Address estrutura de retorno da consulta de CEP
type Address struct {
	State          string `json:"state_short"`
	City           string `json:"city"`
	Neighborhood   string `json:"neighborhood"`
	Street         string `json:"street"`
	IBGE           string `json:"ibge"`
	AdditionalInfo string `json:"additional_info"`
	Bairro         string `json:"bairro"`
}

//NaoPreenchido verifica se existe algum campo preenchido
func (end *Address) NaoPreenchido() bool {
	return end.State == "" && end.City == "" && end.Neighborhood == "" && end.Street == "" && end.IBGE == "" && end.AdditionalInfo == "" && end.Bairro == ""
}
