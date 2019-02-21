#Utiliza a runtime padrão do golang
FROM golang

#Baixa e instala a aplicação
RUN go get github.com/gilliard-okano/dpix

#Configura para o diretório onde foi baixado o código
WORKDIR /go/src/github.com/gilliard-okano/dpix

#Gera o binário da aplicação
RUN go build

#Expoe a porta 8080
EXPOSE 8080

#Configura o entrypoint para o binário
ENTRYPOINT [ "./dpix" ]