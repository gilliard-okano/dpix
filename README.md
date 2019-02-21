# dpix
Processo seletivo Digipix

# Resumo
Aplicação que realiza a consulta de endereço a partir do CEP, utilizando o servidor de homologação da Digipix

# Como rodar o projeto
Clonar o projeto do repositório do Github
```
git clone https://github.com/gilliard-okano/dpix.git
```
Navegar até o diretório do projeto
```
cd dpix
```
Construir e executar a imagem do docker
```
docker build -t dpix . && docker run -p 4000:8080 --rm -t dpix
```
Acessar a URL
```
http://localhost:4000
```