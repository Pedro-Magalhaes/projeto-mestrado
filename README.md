![GitHub Workflow Status](https://img.shields.io/github/workflow/status/Pedro-Magalhaes/projeto-mestrado/Go?label=Monitor%20build)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/Pedro-Magalhaes/projeto-mestrado/Go-Consumer?label=Consumer%20build)
# projeto-mestrado
Repositório para projeto do mestrado


## Projetos com golang
Nesse projeto será implementado um monitor de arquivos em tempo real que usa o kafka para coordenação de réplicas e para enviar trafegar o conteudo dos arquivos entre os monitores e os consumidores.
Foi usado o linux e golang 1.17.3. o linux tem que ter o gcc e g++ para a instalação da lib do kafka

## Execução

Para executar o monito temos que ter o kafka rodando, para isso basta usar o docker-compose da raiz do projeto

```
docker-compose up akhq
```

Com o kafka rodando o monitor pode ser levantando indo para a pasta do monitor:

```
cd monitor/
./run_monitor.sh
```

Para executarmos o consumidor para testes manuais podemos rodar:

```
cd consumer/
./run_consumer.sh
```

### configurações

O monitor pode ser configurado através do arquivo monitor/config.json