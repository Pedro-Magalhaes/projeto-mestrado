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

func Start(config *kafka.ConfigMap, state *SafeMap) (Runnable, error) {
	consumerRoutine, err := NewConsumer(config, state)
	if err != nil {
		fmt.Println("Erro criando consumer. %w\n", err.Error())
		return nil, err
	}
	f := func(cordchan chan bool) {
		consumerChan := make(chan bool)
		go consumerRoutine(consumerChan)

		<-consumerChan // wait consumer
		fmt.Println("Coord will end")
		cordchan <- true // termination
	}
	return f, nil
}
