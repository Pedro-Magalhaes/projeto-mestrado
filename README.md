![GitHub Workflow Status](https://img.shields.io/github/workflow/status/Pedro-Magalhaes/projeto-mestrado/Go?label=Monitor%20build)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/Pedro-Magalhaes/projeto-mestrado/Go-Consumer?label=Consumer%20build)
# projeto-mestrado
Repositório para projeto do mestrado


## Projetos com golang
Nesse projeto será implementado um monitor de arquivos em tempo real que usa o kafka para coordenação de réplicas e para enviar trafegar o conteudo dos arquivos entre os monitores e os consumidores.
Foi usado o linux e golang 1.17.3. o linux tem que ter o gcc e g++ para a instalação da lib do kafka

## Organização

O programa principal fica na pasta "monitor" e o programa auxiliar que serve para testar o monitor fica na pasta "consumer"

No monitor os modulos principais ficam na pasta pkg, a main fica na pasta cmd/monitor/main.go 

Os módulos tem a seguinte funcionalidades:

+ coordinator: Pacote que serve de entrada para a main iniciar o monirtor

+ Config: Pacote que vai carregar a configuração do monitor via arquivo de configuração

+ consumer/consumer: Pacote com a funcionalidade principal do monitor. Fica responsavel por consumir dos tópicos do kafka e adiconar/remover arquivos monitorados, lidar com rebalancing do consumidores, recuperando o 	estado de alguma partição que receber no inicio ou durante a execução

+ consumer/resource_state.go: Pacote que vai implementar uma forma de impedir a concorrencia na hora de manipular o estado 	dos recursos  endo observados. Baseado no padrão: https://gobyexample.com/stateful-goroutines utiliza canais de comunicação e uma única rotina que vai sincronizar os pedidos de acesso à memória

+ consumer/resource_watcher.go: Pacote que implementa a monitoração de arquivos. Vai criar a monitoração e sempre que houver escrita vai enviar os novos bytes escritos para uma callback recebida. Assume que o arquivo sempre sofre "append" ou seja ele só cresce e nada é deletado.

+ consumer/resource.go: Pacote que implementa um objeto do tipo Resource e encapsula sua manipulação

+ producer: Pacote que implementa um singleton para que os outros pacotes possam produzir mensagens para o kafka

+ util/types.go: módulo que contém tipos compartilhados

## Execução

Para executar o monitor temos que ter o kafka rodando, para isso basta usar o docker-compose da raiz do projeto

```
docker-compose up akhq
```

Com o kafka rodando o monitor pode ser levantando indo para a pasta do monitor:

```
cd monitor/
./run_monitor.sh
```
### configurações

O monitor pode ser configurado através do arquivo monitor/config.json

Para executarmos o consumidor para testes manuais podemos rodar:

```
cd consumer/
./run_consumer.sh
```

Por padrão o consumidor sobe na porta 7777, assim se podemos acessar a página http://localhost:7777/?action=# para executarmos o monitor

