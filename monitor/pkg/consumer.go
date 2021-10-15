package coord

import (
	"fmt"
	"strings"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type resourceState struct {
	creatingWatcher bool
	beeingWatched   bool
}

type ResourceSafeMap struct {
	mu          sync.Mutex
	resourceMap map[string]resourceState
}

func callback(v []byte) bool {
	fmt.Println("Callback chamada")
	return true
}

func handleMessage(value []byte, state *ResourceSafeMap) {
	msg := string(value)
	msg = strings.TrimSpace(msg)
	projeto := "ProjetoPlaceholder"
	r, err := CreateResource(msg, projeto)
	if err != nil {
		fmt.Println("Error: could not create a resource is the msg a valid path? Msg: " + msg)
		return
	}

	safeBool := safeBool{work: true}

	state.mu.Lock()
	sm := state.resourceMap[r.GetPath()]

	if sm.beeingWatched || sm.creatingWatcher {
		fmt.Println("Resourece is already being watched")
	} else {
		sm.creatingWatcher = true
		state.resourceMap[r.GetPath()] = sm
		go WatchResource(r, &safeBool, 32, callback, state)
	}
	state.mu.Unlock()

}

func NewConsumer(config *kafka.ConfigMap, state *SafeMap) (Runnable, error) {

	resourcesMap := ResourceSafeMap{resourceMap: make(map[string]resourceState)}
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
