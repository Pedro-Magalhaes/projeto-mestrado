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

type Runnable func(c chan bool)

func Start(config *kafka.ConfigMap, state *SafeMap) (Runnable, error) {
	topic := "monitor_interesse"
	c, err := kafka.NewConsumer(config)
	if err != nil {
		fmt.Println("Erro criando consumer. %w,%w\n", topic, c.String())
		return nil, err
	}
	f := func(cc chan bool) {

		defer c.Close()
		err := c.SubscribeTopics([]string{topic}, rebalance)
		if err != nil {
			fmt.Println("Erro ao conectar", err)
			cc <- false
		} else {
			fmt.Println("Iniciando espera por mensagens")
			for {
				ev := c.Poll(2000)
				switch e := ev.(type) {
				case *kafka.Message:
					fmt.Printf("%% Message on %s:\n%s\n",
						e.TopicPartition, string(e.Value))
				case kafka.PartitionEOF:
					fmt.Printf("%% Reached %v\n", e)
				case kafka.Error:
					fmt.Printf("%% ERROR %v\n", e)
					cc <- false
				default:
					fmt.Printf("Ignored %v\n", e)
				}
			}
		}
		cc <- true // termination
	}
	return f, nil
}
