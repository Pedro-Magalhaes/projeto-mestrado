package consumer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pfsmagalhaes/monitor/pkg/config"
	"github.com/pfsmagalhaes/monitor/pkg/producer"
	"github.com/pfsmagalhaes/monitor/pkg/util"
)

func buildCallback(p producer.Producer, resource *Resource) watchCallback {
	return func(v []byte, offset int64) bool { // TODO: tratar erro

		newMsg := util.FileChunkMsg{Msg: v, Offset: offset, Lenth: len(v)}

		msg, err := json.Marshal(newMsg)
		if err != nil {
			fmt.Println("ERROR: Could not marshal msg!")
			fmt.Println(err.Error())
			return false
		}
		encodedTopic := base64.StdEncoding.EncodeToString([]byte(resource.path))
		p.Write(msg, encodedTopic)
		return true
	}
}

func rebalance(consumer *kafka.Consumer, event kafka.Event) error {
	fmt.Print(consumer.String())
	fmt.Printf("event: %v\n", event)
	return nil
}

func handleMessage(value []byte, state *util.ResourceSafeMap, p producer.Producer, conf *config.Config) {
	file := string(value)
	file = strings.TrimSpace(file)
	if conf.BasePath != "" {
		file = conf.BasePath + file
	}
	projeto := "ProjetoPlaceholder"
	r, err := CreateResource(file, projeto)
	if err != nil {
		fmt.Println("Error: could not create a resource is the msg a valid path? Msg: " + file)
		return
	}

	state.Mu.Lock()
	sm := state.ResourceMap[r.GetPath()]

	if sm.BeeingWatched || sm.CreatingWatcher {
		fmt.Println("Resourece is already being watched")
	} else {
		sm.CreatingWatcher = true
		sm.KeepWorking = true
		state.ResourceMap[r.GetPath()] = sm
		cb := buildCallback(p, r)
		go WatchResource(r, conf.ChunkSize, cb, state)
	}
	state.Mu.Unlock()

}

func NewConsumer(config *kafka.ConfigMap, conf *config.Config, state *util.SafeBoolMap) (util.Runnable, error) {

	resourcesMap := util.ResourceSafeMap{ResourceMap: make(map[string]util.ResourceState)}
	topic := conf.MonitorTopic
	jobInfoTopic := conf.JobInfoTopic
	c, err := kafka.NewConsumer(config)
	p := producer.GetProducer()
	if err != nil {
		fmt.Println("Erro criando consumer. Interesse: %w, jobInfo: %w, consumer: %w\n", topic, jobInfoTopic, c.String())
		return nil, err
	}
	f := func(cc chan bool) {
		defer c.Close()
		err := c.SubscribeTopics([]string{topic, jobInfoTopic}, rebalance)
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
					if strings.Compare(*e.TopicPartition.Topic, topic) == 0 {
						handleMessage(e.Value, &resourcesMap, p, conf)
					} else if strings.Compare(*e.TopicPartition.Topic, jobInfoTopic) == 0 {
						fmt.Printf("Receive job info on topic: %s. message: %s\n", jobInfoTopic, e.Value)
					} else {
						fmt.Println("Unknown topic: %w", e.TopicPartition)
					}
				case kafka.PartitionEOF:
					fmt.Printf("%% Reached %v\n", e)
				case kafka.Error:
					fmt.Printf("%% CONSUMER ERROR %v\n", e)
					cc <- false // Comunica erro e encerra monitor TODO: Como melhorar?
					return
				default: // poll deu timeout.
					// fmt.Printf("Ignored %v\n", e)
				}
			}
		}
	}
	return f, nil
}
