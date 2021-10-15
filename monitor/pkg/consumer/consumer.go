package consumer

import (
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pfsmagalhaes/monitor/pkg/util"
)

func callback(v []byte) bool {
	fmt.Println("Callback chamada")
	return true
}

func rebalance(consumer *kafka.Consumer, event kafka.Event) error {
	fmt.Print(consumer.String())
	fmt.Printf("event: %v\n", event)
	return nil
}

func handleMessage(value []byte, state *util.ResourceSafeMap) {
	msg := string(value)
	msg = strings.TrimSpace(msg)
	projeto := "ProjetoPlaceholder"
	r, err := CreateResource(msg, projeto)
	if err != nil {
		fmt.Println("Error: could not create a resource is the msg a valid path? Msg: " + msg)
		return
	}

	safeBool := safeBool{work: true}

	state.Mu.Lock()
	sm := state.ResourceMap[r.GetPath()]

	if sm.BeeingWatched || sm.CreatingWatcher {
		fmt.Println("Resourece is already being watched")
	} else {
		sm.CreatingWatcher = true
		state.ResourceMap[r.GetPath()] = sm
		go WatchResource(r, &safeBool, 32, callback, state)
	}
	state.Mu.Unlock()

}

func NewConsumer(config *kafka.ConfigMap, state *util.SafeBoolMap) (util.Runnable, error) {

	resourcesMap := util.ResourceSafeMap{ResourceMap: make(map[string]util.ResourceState)}
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
			return
		}
		fmt.Println("Iniciando espera por mensagens")
		for {
			select {
			case <-cc:
				fmt.Println("consumer channel closed exiting")
				return
			default:
				ev := c.Poll(2000)
				switch e := ev.(type) {
				case *kafka.Message:
					fmt.Printf("%% Message on %s:\n%s\n",
						e.TopicPartition, string(e.Value))
					handleMessage(e.Value, &resourcesMap)
				case kafka.PartitionEOF:
					fmt.Printf("%% Reached %v\n", e)
				case kafka.Error:
					fmt.Printf("%% CONSUMER ERROR %v\n", e)
					cc <- false
					return
				default:
					fmt.Printf("Ignored %v\n", e)
				}
			}
		}
	}
	return f, nil
}
