package coord

import (
	"fmt"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Behaviour func(kafka.Message)

// SafeCounter is safe to use concurrently.
type SafeMap struct {
	mu sync.Mutex
	V  map[string]bool
}

func rebalance(consumer *kafka.Consumer, event kafka.Event) error {
	fmt.Print(consumer.String())
	fmt.Printf("event: %v\n", event)
	return nil
}

type Runnable func(chan bool)

func Create(config *kafka.ConfigMap, state *SafeMap) (Runnable, error) {
	// activeResources := util.NewSafeBoolMap()
	consumerRoutine, err := NewConsumer(config, state)
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
		// <-consumerChan // wait consumer
	}
	return f, nil
}
