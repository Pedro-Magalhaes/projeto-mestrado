package coord

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pfsmagalhaes/monitor/pkg/config"
	"github.com/pfsmagalhaes/monitor/pkg/consumer"
	"github.com/pfsmagalhaes/monitor/pkg/util"
)

type Behaviour func(kafka.Message)

// SafeCounter is safe to use concurrently.

func Create(config *kafka.ConfigMap, state *util.SafeBoolMap, conf *config.Config) (util.Runnable, error) {
	// activeResources := util.NewSafeBoolMap()
	consumerRoutine, err := consumer.NewConsumer(config, state)
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
