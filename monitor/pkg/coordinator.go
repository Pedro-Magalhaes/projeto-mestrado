/*
	Pacote que serve de entrada para a main iniciar o monirtor

	Autor: Pedro Magalhães
*/
package coord

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pfsmagalhaes/monitor/pkg/config"
	"github.com/pfsmagalhaes/monitor/pkg/consumer"
	"github.com/pfsmagalhaes/monitor/pkg/util"
)

/*
	Função que retorna a rotina principal do programa e fica observando o canal
	da main e do consumer para ver se precisa para o programa
	Recebe um kafka configMap e a configuração do monitor. Retorna uma função
	do tipo util.Runnable ou um erro
*/
func Create(kConfig *kafka.ConfigMap, conf config.Config) (util.Runnable, error) {
	// activeResources := util.NewSafeBoolMap()
	consumerRoutine, err := consumer.NewConsumer(kConfig, conf)
	if err != nil {
		fmt.Println("Erro criando consumer. %w\n", err.Error())
		return nil, err
	}
	f := func(coordchan chan bool) {
		consumerChan := make(chan bool)
		go consumerRoutine(consumerChan)

		select {
		case <-consumerChan:
			fmt.Println("Coord will end. Got finish message from consumers")
			coordchan <- true // termination
		case <-coordchan:
			fmt.Println("Coord will end. Got finish message from main")
			close(consumerChan)
		}
	}
	return f, nil
}
