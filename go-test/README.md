# testes em Golang

Projeto em Go para criar um programa para facilitar a criação de testes para o monitor com comunicação com Kafka e docker

## Dependencias

    golang 1.17+
	github.com/confluentinc/confluent-kafka-go v1.8.2
	github.com/fsouza/go-dockerclient v1.7.11


## Rodando o teste

O arquivo principal que está sendo usado para testar o desenvolvimento é o [test.go](./test.go) para executa-lo basta usar o comando: 

```
go run test.go
```