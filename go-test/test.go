package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pfsmagalhaes/go-test/stage"
	"github.com/pfsmagalhaes/go-test/topic"
)

type FileChunkMsg struct {
	Msg    string `json:"msg"`
	Offset int64  `json:"offset"`
	Lenth  int    `json:"lenth"`
}

var client *docker.Client
var container *docker.Container
var err error
var kafkaAdmClient *kafka.AdminClient
var kafkaProducer *kafka.Producer
var kafkaConsumer *kafka.Consumer
var messageMutex sync.Mutex
var kafkaMessages map[string][]string
var linhasArquivo1 []string

func internalSetup() {
	// AdmClient para criar topicos
	kafkaAdmClient, err = kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092",
		"acks": "all"})
	if err != nil {
		fmt.Println("Erro criando adm do kafka")
		panic(err)
	}

	// Producer para mandar mensagens
	kafkaProducer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092",
		"acks": "all"})
	if err != nil {
		fmt.Println("Erro criando producer do kafka")
		panic(err)
	}

	// Consumer do kafka
	kafkaConsumer, err = kafka.NewConsumer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092", "group.id": "consumer-teste-id",
		"acks": "all"})
	if err != nil {
		fmt.Println("Erro criando consumer do kafka")
		panic(err)
	}

	// Docker client para executar imagens
	client, err = docker.NewClientFromEnv()
	if err != nil {
		fmt.Println("Erro criado cliente do kafka")
		panic(err)
	}
	kafkaMessages = make(map[string][]string)
	linhasArquivo1 = []string{"Primeira linha\n", "Segunda linha\n", "Ultima linha"}
}

func setup() {
	var t, e = topic.LoadTopicConfig("topics.json")
	if e != nil {
		fmt.Println("Erro ao importar arquivo json de topicos")
		panic(e)
	}
	createTopics(t)
}

func addKafkaMessage(topic, message string) {
	messageMutex.Lock()
	if kafkaMessages[topic] == nil {
		kafkaMessages[topic] = make([]string, 0)
	}
	kafkaMessages[topic] = append(kafkaMessages[topic], message)
	messageMutex.Unlock()
}

func getKafkaLastMessage(topic string) string {
	return getKafkaMessage(topic, len(kafkaMessages[topic]))
}

func getKafkaMessage(topic string, pos int) string {
	messageMutex.Lock()
	message := kafkaMessages[topic][pos]
	messageMutex.Unlock()
	return message
}

func getKafkaMessagesArray(topic string) []string {
	var messages []string
	messageMutex.Lock()
	if kafkaMessages[topic] != nil {
		messages = kafkaMessages[topic]

	}
	messageMutex.Unlock()
	return messages
}

func handleMessages() {
	// Process messages
	for {
		ev, err := kafkaConsumer.ReadMessage(100 * time.Millisecond)
		if err != nil {
			// Errors are informational and automatically handled by the consumer
			continue
		}
		// vou ignorar as keys por enquanto
		fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
			*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
		addKafkaMessage(*ev.TopicPartition.Topic, string(ev.Value))
	}
}

func startConsumer(topics []string) {
	println("Iniciando consumer com os seguintes topicos: ", topics)
	err := kafkaConsumer.SubscribeTopics(topics, nil)
	if err != nil {
		fmt.Println("Erro subscrevendo aos topicos")
		panic(err)
	}
	go handleMessages()
}

func createTopics(t *topic.TopicConfig) {
	var topicNames []string = make([]string, len(t.Topics))
	var specifications []kafka.TopicSpecification = make([]kafka.TopicSpecification, len(t.Topics))
	for i, v := range t.Topics {
		topicNames[i] = v.Name
		specifications[i] = kafka.TopicSpecification{
			Topic:             v.Name,
			NumPartitions:     v.NumPartitions,
			ReplicationFactor: 1,
		}
	}
	var _, e = kafkaAdmClient.DeleteTopics(context.Background(), topicNames)
	if e != nil {
		fmt.Println("Não foi possivel deletar os topicos")
		fmt.Println(topicNames)
		panic(e)
	}

	// TODO: como aguardar a deleção dos tópicos?
	time.Sleep(time.Second * 1)

	var topics, err = kafkaAdmClient.CreateTopics(context.Background(), specifications)
	if err != nil {
		fmt.Println("Não foi possivel criar os tópicos")
		fmt.Println(topicNames)
		panic(err)
	}

	fmt.Printf("topics: %v\n", topics)

	// TODO: como ver se o tópico está pronto para receber mensagens?
	time.Sleep(time.Second * 1)

	for _, v := range t.Topics {
		if len(v.Messages) > 0 {
			fmt.Println("Deveria colocar as mensagens: ")
			fmt.Println(v.Messages)
		}
		for _, m := range v.Messages {
			kafkaProducer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &v.Name, Partition: int32(m.Partition)},
				Value:          []byte(m.Message),
			}, nil)
		}
	}

}

// stage 1
// escreve no arquivo(append), declara interesse
func writeToFile(fileName string, text string) {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		println("Erro ao abrir arquivo", fileName)
		panic(err)
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		println("Erro ao escrever no arquivo", fileName)
		panic(err)
	}
}

func observeFile(fileName string) {
	topicName := "monitor_interesse"
	deliveryChan := make(chan kafka.Event)
	kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
		Value:          []byte(fmt.Sprintf(`{"watch": true, "path": "%s", "project": "p1" }`, fileName)),
		Key:            []byte("job01"),
	}, deliveryChan)
	ev := <-deliveryChan // vai retornar só após receber a info de que a mensagem foi enviada
	fmt.Printf("Mensagem do canal de interesse. \nev: %v\n", ev.String())
}

// TODO: Como melhorar mecanismo do for infinito aguardando a posição esperada ser preenchida?
func checkMessageReceivedWithTimeout(topic string, pos int, expected string, timeout time.Duration) {
	c := make(chan bool, 1)
	stop := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				time.Sleep(time.Second * 2)
				arr := getKafkaMessagesArray(topic)
				if pos >= len(arr) { // array menor que a posição recebida
					println("DEBUG: soft Timeout na checagem do array de mensagns kafka. pos, len", pos, len(arr))
					continue
				} else {
					msg := arr[pos]
					var fileChunk FileChunkMsg
					err := json.Unmarshal([]byte(msg), &fileChunk)
					if err != nil {
						c <- false
						panic(err)
					}

					if strings.Compare(fileChunk.Msg, expected) == 0 {
						c <- true
					} else {
						println("Expected:", expected, "Got:", fileChunk.Msg)
						c <- false
					}
				}
			}
		}
	}()
	select {
	case res := <-c:
		println("Resultado da checagem", res)
	case <-time.After(timeout):
		println("Timeout na checagem de mensagens")
	}
	close(stop)
}

func writeToFileFabric(fileName, text string) func(chan bool) {
	return func(c chan bool) {
		println("DEBUG: writeToFile")
		writeToFile(fileName, text)
	}
}

func continuousWritingFabric(fileName, text string, d time.Duration) func(chan bool) {
	return func(c chan bool) {
		for {
			select {
			case <-c:
				log.Default().Println("Stopping continuousWriting")
				return
			case <-time.After(d):
				writeToFile(fileName, text)
			}
		}
	}
}

func observeFileFabric(fileName string, d time.Duration) func(chan bool) {
	return func(c chan bool) {
		println("DEBUG: observeFile chamado")
		observeFile(fileName)
		time.Sleep(d) //espera tres segundos pro monitor observar o arquivo?
		println("Observe finalizado")
	}
}

func checkMessageReceivedWithTimeoutFabric(topic string, pos int, expected string, timeout time.Duration) func(chan bool) {
	return func(c chan bool) {
		println("DEBUG: checkMessageReceivedWithTimeout chamado")
		checkMessageReceivedWithTimeout(topic, pos, expected, timeout)
	}
}

func checkMessageReceivedWithTimeoutStage2(c chan bool) {
	println("DEBUG: checkMessageReceivedWithTimeoutStage2")
	checkMessageReceivedWithTimeout("test_files__test_file_0.txt", 0, linhasArquivo1[0], time.Second*30)
}

// fazer um arrray de WG para que cada job  seja inserido no estagio que ele vai terminar
func main() {
	l := log.Default()
	var e error
	topics := []string{"test_files__test_file_0.txt", "test_files__test_file_1.txt"}

	defer teardown()
	internalSetup()
	setup()
	startConsumer(topics) // cria uma rotina para receber as mensagens e preencher um mapa [topico] -> [mensagens]

	config := docker.Config{
		Image:        "monitor:0.0.1-snapshot",
		OpenStdin:    true,
		StdinOnce:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		// Healthcheck: ,
	}
	hostConfig := docker.HostConfig{
		RestartPolicy: docker.RestartUnlessStopped(),
		Mounts: []docker.HostMount{
			{
				Type:   "bind",
				Source: "/local/ProgramasLocais/Documents/pessoal/Puc/mestrado/projeto-mestrado/go-test/test_files",
				Target: "/go/src/app/test_files",
			},
			{
				Type:   "bind",
				Source: "/local/ProgramasLocais/Documents/pessoal/Puc/mestrado/projeto-mestrado/go-test/config-monitor.json",
				Target: "/go/src/app/config.json",
			},
		},
	}
	networkConfig := docker.NetworkingConfig{EndpointsConfig: map[string]*docker.EndpointConfig{"projeto-mestrado_default": {}}}
	dockerCreatorConfig := docker.CreateContainerOptions{Name: "monitor_1_teste", Config: &config, HostConfig: &hostConfig, NetworkingConfig: &networkConfig}

	container, e = client.CreateContainer(dockerCreatorConfig)
	if e != nil {
		l.Printf("e: %v\n", e)
	} else {
		l.Printf("c: %v\n", container)

		cerr := client.StartContainer(container.ID, container.HostConfig)

		// olhar healthcheck

		if cerr != nil {
			l.Printf("cerr: %v\n", cerr)
			return
		}
		time.Sleep(time.Second * 7) // esperando para iniciar

		file0 := "test_files/test_file_0.txt"
		file1 := "test_files/test_file_1.txt"
		l.Println("Iniciando testes\n")
		// stages := {
		// id: 1
		// 	jobs: {
		// 		{ funcao: observeFileFabric(file0, 1*time.Second)},
		// 		{ funcao: observeFileFabric(file1, 1*time.Second)}
		// 		{funcao: observeFileFabric(file1, 1*time.Second), endId: 3}
		// 	}
		// }
		st1 := stage.CreateStage(1)
		st1.AddJob(observeFileFabric(file0, 1*time.Second))
		st1.AddJob(observeFileFabric(file0, 25*time.Second))
		st1.AddJob(writeToFileFabric(file0, linhasArquivo1[0]))
		st2 := stage.CreateStage(2)
		st2.AddJob(checkMessageReceivedWithTimeoutFabric("test_files__test_file_0.txt", 0, linhasArquivo1[0], time.Second*30))
		st3 := stage.CreateStage(3)
		st3.AddJob(func(c chan bool) {
			log.Default().Println("Job do stage 3")
		})
		st1.AddJob2(continuousWritingFabric(file1, "Mais uma linha\n", time.Second), st2)
		l.Println("Rodando stage 1")
		st1.Run()
		l.Println("Rodando stage 2")
		st2.Run()
		l.Println("Rodando stage 3")
		st3.Run()
		// wg1 := stage1()
		// wg1.Wait()
		// wg2 := stage2()
		// wg2.Wait()

		// wg3 := stage3()
		// wg3.Wait()
		l.Println("testes finalizados")
	}
}

func remove(client *docker.Client, container *docker.Container) {
	fmt.Printf("parando container: %v\n", container.ID)
	if cerr := client.StopContainer(container.ID, 10); cerr != nil {
		fmt.Printf("could not stop container: %v\n", cerr)
	}
	if err := client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID}); err != nil {
		fmt.Printf("err: %v\n", err)
	}
}

func teardown() {

	if container != nil {
		fmt.Printf("removendo container: %v\n", container.ID)
		remove(client, container)
	} else {
		fmt.Printf("container nulo: %v\n", container)
	}
	if kafkaAdmClient != nil {
		println("removendo kafka Adm")
		kafkaAdmClient.Close()
	}
	if kafkaProducer != nil {
		println("removendo kafka Producer")
		kafkaProducer.Close()
	}
	if kafkaConsumer != nil {
		println("removendo kafka Consumer")
		kafkaConsumer.Close()
	}

}

// WG.add(t)

// <-WG

// asdada
