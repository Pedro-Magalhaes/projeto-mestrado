/*
	Pacote com a funcionalidade principal do monitor. Fica responsavel por consumir dos tópicos do
	kafka e adiconar/remover arquivos monitorados, lidar com rebalancing do consumidores, recuperando o
	estado de alguma partição que receber no inicio ou durante a execução

	Autor: Pedro Magalhães
*/
package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pfsmagalhaes/monitor/pkg/config"
	"github.com/pfsmagalhaes/monitor/pkg/producer"
	"github.com/pfsmagalhaes/monitor/pkg/util"
)

/*
	Cria a callback que vai chamar o producer para escrever os dados recebidos pelo watcher
	que está observando um arquivo. Essa callback vai ser passada para o resource_watcher
	Recebe o producer e a referencia para o recurso sendo observado e retorna a função que
	escreve os dados
*/
func buildCallback(p producer.Producer, resource *Resource) watchCallback {
	return func(v []byte, offset int64) bool { // TODO: tratar erro

		newMsg := util.FileChunkMsg{Msg: string(v), Offset: offset, Lenth: len(v)}

		msg, err := json.Marshal(newMsg)
		if err != nil {
			fmt.Println("ERROR: Could not marshal msg!")
			fmt.Println(err.Error())
			return false
		}
		topic := resource.GetResourceTopic()
		p.Write(msg, topic, "")
		return true
	}
}

/*
	Função que vai tratar o rebalance do grupo de consumidores do tópico que contém as mensagens indicando os
	recursos que devem ser obsevados. Temos dois eventos que serão tratados: AssignedPartitions e RevokedPartitions
	Quando uma patição é recebida o monitor vai buscar o estado dessa partição no tópico de estado, se ele existir,
	e no caso de revoked o monitor vai parar os recursos vinculados as partições
*/
func rebalance(consumer *kafka.Consumer, event kafka.Event) error {
	c, _ := config.GetConfig("")
	if e, ok := event.(kafka.AssignedPartitions); ok {
		parts := make(map[int32]bool, len(e.Partitions))
		for _, p := range e.Partitions {
			parts[p.Partition] = true
		}
		log.Printf("PARTS RECEIVED: Parts: %+v\nPartitions: %+v\n", parts, e.Partitions)
		resourceChan := make(chan []Resource, len(parts)) // buffered to avoid blocking
		tBefore := time.Now()
		getLastMsgFromTopicPartition("monitor_estado", parts, resourceChan)

		for i := 0; i < len(parts); i++ {
			recoverState(<-resourceChan)
		}
		log.Printf("TIME SPENT WAITING FOR STATE RECOVERY: %+v. Segundos: %f\n", time.Since(tBefore), time.Since(tBefore).Seconds())
	} else if e, ok := event.(kafka.RevokedPartitions); ok {
		for _, p := range e.Partitions {
			if *p.Topic == c.MonitorTopic {
				log.Printf("Revoking partition %+v\n", p)
				DeletePartition(p.Partition)
			}
		}

	} else {
		fmt.Printf("REBALANCE ERROR. Unknow type: %#v\n", e)
	}
	return nil
}

/*
	Recria os watchers para a lista de recursos recebidos. Não testa para ver se os recursos já estão sendo
	observados porque só é chamada na recuperação de estado no AssignedPartitions.
	Recebe uma lista de Resoureces
*/
func recoverState(resources []Resource) {
	for _, r := range resources {
		if r.Jobid == "" || r.Path == "" || r.Project == "" {
			log.Println("Recovery skipped for resource: ", r)
			continue
		}
		channel := make(chan bool)
		resourceState := &ResourceState{CreatingWatcher: true, BeeingWatched: false, KeepWorking: &channel, R: &r}
		createWatcherForResource(&r, resourceState)
	}
}

/*
	Função que vai recuperar o estado de cada partição pegando a última mensagem enviada para o tópico
	de "estado" na partição correspondente.
	Recebe: O nome do tópico, um mapa hash com a chave sendo o id da partição e um boleano indicando se
	o estado da partição deve ser recuperado e o canal para onde deve enviar o estado de cada partição
	A função fará um numero fixo de tentativas de recuperar o estado das partições se não conseguir vai
	enviar uma lista nula para o canal. Isso vai ocorrer quando não ouver mensagens na partição ou ocorrer
	algum problema de rede.
*/
func getLastMsgFromTopicPartition(t string, p map[int32]bool, resourceChan chan []Resource) {
	conf, _ := config.GetConfig("")
	kConf := kafka.ConfigMap{
		"bootstrap.servers":               conf.KafkaUrl,
		"group.id":                        time.Now().String(),
		"go.application.rebalance.enable": true,
		"auto.offset.reset":               "earliest"}

	c, err := createConsumer(&kConf)
	if err != nil {
		log.Println("Could not load create the consumer for watcher")
		panic(err)
	}
	defer c.Close()
	c.Subscribe(t, func(c *kafka.Consumer, e kafka.Event) error {
		if ev, ok := e.(kafka.AssignedPartitions); ok {
			log.Printf("Assinged MONITOR_ESTADO>>> %#v\n", ev.Partitions)
			parts := make([]kafka.TopicPartition, 0)
			for _, tp := range ev.Partitions {
				if p[tp.Partition] {
					tp.Offset = kafka.OffsetTail(1) // Set start offset to 1 messages from end of partition
					parts = append(parts, tp)
				}
			}
			log.Printf("Assign %v\n", parts)
			c.Assign(parts)
		}
		return nil
	})
	if err != nil {
		log.Println("wathcer Could not load subscribe to")
		panic(err)
	}
	log.Println("GET STATE subscribed to ", t)
	pCount := 0

	for i := 0; i < 20; i++ {
		if pCount == len(p) {
			break
		}
		ev := c.Poll(1000)
		switch e := ev.(type) {
		case kafka.AssignedPartitions:
			parts := make([]kafka.TopicPartition,
				len(e.Partitions))
			for i, tp := range e.Partitions {
				tp.Offset = kafka.OffsetTail(1) // Set start offset to 1 messages from end of partition
				parts[i] = tp
			}
			fmt.Printf("Assign %v\n", parts)
			c.Assign(parts)
		case *kafka.Message:
			fmt.Printf("***MSG %v\n", e.TopicPartition)
			m := []Resource{}
			ee := json.Unmarshal(e.Value, &m)
			if ee != nil {
				log.Println("Getting offset: ERROR unmarshal kafka msg", ee)
			}
			log.Printf("***** Got Response on partition %d: %+v\n", e.TopicPartition.Partition, m)
			pCount += 1
			resourceChan <- m
		case kafka.PartitionEOF:
			fmt.Printf("%% Reached %v\n", e)
		case kafka.Error:
			log.Println("Kafka error when trying to recover monitor state", e)

		default:
			// fmt.Printf("Ignored %v\n", e)
		}
	}
	for i := pCount; i < len(p); i++ {
		resourceChan <- nil
	}
}

/*
	Função que verifica o último offset enviado no tópico que é usado para enviar o conteudo dos arquivos
	observados. Se o tópico não existir ou o timeout acabar o recurso começa a ser observado com offset 0
	O timeout é grande o suficiente para que se o tópico existir, seja possível recuperar a última mensagem
	do tópico e obter o offset
*/
func getConsumerOffset(t string) int64 {
	conf, _ := config.GetConfig("")
	kConf := kafka.ConfigMap{
		"bootstrap.servers":               conf.KafkaUrl,
		"group.id":                        time.Now().String(),
		"go.application.rebalance.enable": true,
		"auto.offset.reset":               "earliest"}

	c, err := createConsumer(&kConf)
	if err != nil {
		log.Println("Could not load create the consumer for watcher")
		panic(err)
	}
	defer c.Close()
	c.Subscribe(t, nil)
	if err != nil {
		log.Println("wathcer Could not load subscribe to")
		panic(err)
	}
	log.Println("getoffset subscribed to ", t)
	off := int64(0)
	first := true
	for i := 0; i < 20; i++ {
		ev := c.Poll(1000)
		if ev == nil {
			if off != 0 {
				return off
			}
			continue
		}
		switch e := ev.(type) {
		case kafka.AssignedPartitions:
			parts := make([]kafka.TopicPartition,
				len(e.Partitions))
			for i, tp := range e.Partitions {
				tp.Offset = kafka.OffsetTail(1) // Set start offset to 1 messages from end of partition
				parts[i] = tp
			}
			fmt.Printf("Assign %v\n", parts)
			c.Assign(parts)
		case *kafka.Message:
			if first {
				first = false
				e.TopicPartition.Offset = kafka.OffsetTail(1)
				c.Assign([]kafka.TopicPartition{e.TopicPartition})
			}
			m := util.FileChunkMsg{}
			ee := json.Unmarshal(e.Value, &m)
			if ee != nil {
				log.Println("Getting offset: ERROR unmarshal kafka msg", ee)
			}
			off = m.Offset + int64(m.Lenth)
		case kafka.PartitionEOF:
			fmt.Printf("%% Reached %v\n", e)
			return off
		case kafka.Error:
			log.Println("get offset Kafka error returning 0.", e)
			return off
		default:
			// fmt.Printf("Ignored %v\n", e)
		}
	}
	return off
}

/*
	Função responsavel por criar a rotina que observa os arquivos.
*/
func createWatcherForResource(r *Resource, rs *ResourceState) {
	p := producer.GetProducer()
	c, _ := config.GetConfig("") // ignoring error cause at this point the config should have been accessed multiple times
	cb := buildCallback(p, r)
	r.PutStateToStateStore(rs)
	r.SetCurrentOffset(getConsumerOffset(r.GetResourceTopic()))

	go WatchResource(r, c.ChunkSize, cb, *rs)
}

/*
	Função que escreve no tópico de estado dos monitores o estado atual de uma partição.
	Recebe o id de uma partição, o producer e o nome tópico usado para guardar o estado.
*/
func writeNewPartitionState(partition int32, p producer.Producer, StateTopic string) {
	resources := GetPartitionResources(partition)
	var v []byte
	var e error
	if len(resources) < 1 {
		log.Println("ERROR writeNewPartitionState: LEN OF RESOURCES < 1")
		log.Printf("RESOURCES: %+v\n", resources)
		v, e = json.Marshal("[]")
		if e != nil {
			log.Println("Error - writeNewPartitionState: Could not marshal empty resource!", e)
		}
	} else {
		v, e = json.Marshal(resources)
		if e != nil {
			log.Println("Error - writeNewPartitionState: Could not marshal resources!", e)
		}
	}

	p.WriteToPartition(v, nil, StateTopic, partition)
}

/*
	Função que trata mensagens recebidas no tópico de jobs. Por enquanto esse tópico é um mock e só recebe
	mensagens para indicar que um job terminou para que o monitor pare todos os recursos vinculados à ele
	por isso a função apenas verifica se no objeto recebido existe a string "finshed" e escreve no tópico
	interesse a mensagem para parar os recursos observados vinculados ao job
*/
func handleJobStateMessage(key, value []byte, p producer.Producer) {
	v := string(value)
	if strings.Contains(v, "finished") {
		msg := util.InfoMsg{Path: "", Watch: false, Project: ""} // convention: when path empty and watch == false should stop all from job
		v, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("Error writing msg to sop job", err)
		} else {
			p.Write(v, "monitor_interesse", string(key))
		}
	} else {
		log.Println("Job not finished. nothing to do.")
	}
}

/*
	Função responsável por tratar as mensagens recebidas no tópico que indica os recursos que devem ser
	observados. Ela vai tratar a mensagem e verificar se é preciso começar um novo watcher para um determinado
	recurso ou para-lo.
	Recebe o id da partição do kafka, sua chave, o contúdo, o produtor e a configuração do monitor.
*/
func handleMonitorMessage(partition int32, key, value []byte, p producer.Producer, conf config.Config) {
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
		DeleteJob(partition, jobid)
		writeNewPartitionState(partition, p, conf.StateTopic)
		return
	}
	if conf.BasePath != "" { // Fixme: BasePath não deve ficar aqui. Melhor criar o recurso e usar o basePath na hora de iniciar o watcher
		file = conf.BasePath + file
	}
	r, err := CreateResource(file, msg.Project, jobid, partition)
	if err != nil {
		fmt.Println("Error: could not create a resource is the msg a valid path? Msg: " + file)
		return
	}

	sm := r.GetStateFromStateStore()

	if sm == nil { // creating the first obj
		workChan := make(chan bool)
		sm = &ResourceState{CreatingWatcher: false, BeeingWatched: false, KeepWorking: &workChan, R: r}
		PutResource(partition, jobid, r.GetPath(), sm)
		log.Println("Putting resource since it was nil. New resource:")
		log.Printf("%+v\n", sm)
	} else {
		log.Println("Resource not new:")
		log.Printf("%+v\n", sm)
	}

	if !msg.Watch {
		fmt.Println("Stoping resource  watcher")
		if sm.KeepWorking != nil && (sm.BeeingWatched || sm.CreatingWatcher) {
			close(*sm.KeepWorking)
			resources := GetPartitionResources(partition)
			foundIndex := -1
			for i, v := range resources {
				if v.GetPath() == r.GetPath() {
					foundIndex = i
					break
				}
			}
			resources = append(resources[:foundIndex], resources[foundIndex+1:]...)
			v, e := json.Marshal(resources)
			if e != nil {
				log.Println("ERROR: Could not marshal resources!", e)
			}

			p.WriteToPartition(v, key, conf.StateTopic, partition)
		} else {
			fmt.Println("Resourse was not been watched")
			fmt.Println(sm)
		}
	} else {
		if sm.BeeingWatched || sm.CreatingWatcher {
			fmt.Println("Resource is already being watched")
		} else {
			sm.BeeingWatched = false
			sm.CreatingWatcher = true
			createWatcherForResource(r, sm)

			writeNewPartitionState(partition, p, conf.StateTopic)
		}
	}
}

/*
	Cria um novo consumidor do kafka com o configMap recbido
*/
func createConsumer(c *kafka.ConfigMap) (*kafka.Consumer, error) {
	return kafka.NewConsumer(c)
}

/*
	Função que retorna a função principal que vai rodar o monitor. Ela retorna uma função
	que recebe um canal de comunicação para que possa ser parado por quem instanciar o consumidor
*/
func NewConsumer(kConfig *kafka.ConfigMap, conf config.Config) (util.Runnable, error) {

	monitorTopic := conf.MonitorTopic
	jobInfoTopic := conf.JobInfoTopic
	kConfig.SetKey("max.poll.interval.ms", int(conf.ConsumerTimeout))
	kConfig.SetKey("session.timeout.ms", int(conf.ConsumerSessionTimeout))

	c, err := createConsumer(kConfig)
	p := producer.GetProducer()
	if err != nil {
		fmt.Printf("Erro criando consumer. Interesse: %s, jobInfo: %s, consumer: %v\n", monitorTopic, jobInfoTopic, c)
		return nil, err
	}

	consumerRoutine := func(cc chan bool) {
		defer c.Close()
		defer p.Close()
		err := c.SubscribeTopics([]string{monitorTopic, jobInfoTopic}, rebalance)
		if err != nil {
			fmt.Println("Erro ao conectar", err)
			cc <- false
			return
		}
		// Routine that will sync resources access
		StartStateStore()
		log.Println("Consumer: Store started")
		defer StopStateStore()

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
					if strings.Compare(*e.TopicPartition.Topic, monitorTopic) == 0 {
						handleMonitorMessage(e.TopicPartition.Partition, e.Key, e.Value, p, conf)
					} else if strings.Compare(*e.TopicPartition.Topic, jobInfoTopic) == 0 {
						fmt.Printf("Receive job info on topic: %s. message: %s\n", jobInfoTopic, e.Value)
						handleJobStateMessage(e.Key, e.Value, p)
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
	return consumerRoutine, nil
}
