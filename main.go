package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", ServicoDeEndereco)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
