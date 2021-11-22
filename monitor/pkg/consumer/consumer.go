package consumer

import (
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

		newMsg := util.FileChunkMsg{Msg: string(v), Offset: offset, Lenth: len(v)}

		msg, err := json.Marshal(newMsg)
		if err != nil {
			fmt.Println("ERROR: Could not marshal msg!")
			fmt.Println(err.Error())
			return false
		}
		pathArray := strings.SplitN(resource.path, "/", 1)
		index := 1
		if len(pathArray) < 2 { // no basepath
			index = 0
		}
		topic := strings.ReplaceAll(pathArray[index], "/", "__")
		//encodedTopic := strings.ReplaceAll(resource.path, "/", "__")
		p.Write(msg, topic, "")
		return true
	}
}

// TODO: Implementar logica para recuperar o estado de uma partição
func rebalance(consumer *kafka.Consumer, event kafka.Event) error {
	fmt.Print(consumer.String())
	fmt.Printf("event: %v\n", event)
	return nil
}

// Will handle msgs on the jobstate topic. When a job finish it will send a msg to the monitor_interesse topic
func handleJobStateMessage(key, value []byte, state *util.ResourceSafeMap, p producer.Producer, conf *config.Config) {
	msg := util.InfoMsg{Path: "", Watch: false, Project: ""} // convetion: when path empty and watch == false should stop all from job
	v, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error writing msg to sop job", err)
	} else {
		p.Write(v, "monitor_interesse", string(key))
	}
}

func handleMessage(key, value []byte, state *util.ResourceSafeMap, p producer.Producer, conf *config.Config) {
	jobid := string(key)
	msg := util.InfoMsg{}
	if err := json.Unmarshal(value, &msg); err != nil {
		fmt.Println("Error msg does not respect msg interface!")
		return
	}

	file := msg.Path
	file = strings.TrimSpace(file)
	if !msg.Watch && file == "" {
		fmt.Println("Should stop all watcher from Job: ", string(key))
		fmt.Println("Not implemented!")
		return
	}
	if conf.BasePath != "" { // Fixme: BasePath não deve ficar aqui. Melhor criar o recurso e usar o basePath na hora de iniciar o watcher
		file = conf.BasePath + file
	}
	r, err := CreateResource(file, msg.Project, jobid)
	if err != nil {
		fmt.Println("Error: could not create a resource is the msg a valid path? Msg: " + file)
		return
	}

	state.Mu.Lock()
	sm := state.ResourceMap[r.GetPath()]

	if sm == nil { // creating the first obj
		sm = &util.ResourceState{CreatingWatcher: false, BeeingWatched: false, KeepWorking: make(chan bool)}
		state.ResourceMap[r.GetPath()] = sm
	}

	if !msg.Watch {
		fmt.Println("Stoping resource  watcher")
		if sm.KeepWorking != nil && (sm.BeeingWatched || sm.CreatingWatcher) {
			close(sm.KeepWorking)
		} else {
			fmt.Println("Resourse was not been watched")
		}
	} else {
		if sm.BeeingWatched || sm.CreatingWatcher {
			fmt.Println("Resource is already being watched")
		} else {
			println("consumer state", state)
			println("consumer resouce", r.GetPath())
			cb := buildCallback(p, r)
			sm = &util.ResourceState{CreatingWatcher: true, BeeingWatched: false, KeepWorking: make(chan bool)}
			state.ResourceMap[r.GetPath()] = sm
			go WatchResource(r, conf.ChunkSize, cb, state)
		}
	}
	state.Mu.Unlock()

}

func createConsumer(c *kafka.ConfigMap) (*kafka.Consumer, error) {
	return kafka.NewConsumer(c)
}

func NewConsumer(kConfig *kafka.ConfigMap, conf *config.Config) (util.Runnable, error) {

	resourcesMap := util.ResourceSafeMap{ResourceMap: make(map[string]*util.ResourceState)}
	topic := conf.MonitorTopic
	jobInfoTopic := conf.JobInfoTopic
	c, err := createConsumer(kConfig)
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
						handleMessage(e.Key, e.Value, &resourcesMap, p, conf)
					} else if strings.Compare(*e.TopicPartition.Topic, jobInfoTopic) == 0 {
						handleJobStateMessage(e.Key, e.Value, &resourcesMap, p, conf)
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
