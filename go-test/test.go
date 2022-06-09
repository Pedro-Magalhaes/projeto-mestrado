package main

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pfsmagalhaes/go-test/topic"
)

var client *docker.Client
var container *docker.Container
var err error
var kafkaAdmClient *kafka.AdminClient
var kafkaProducer *kafka.Producer
var stageChan chan int

func internalSetup() {
	kafkaAdmClient, err = kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092",
		"acks": "all"})
	if err != nil {
		fmt.Println("Erro criando adm do kafka")
		panic(err)
	}

	kafkaProducer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092",
		"acks": "all"})
	if err != nil {
		fmt.Println("Erro criando producer do kafka")
		panic(err)
	}
	client, err = docker.NewClientFromEnv()
	if err != nil {
		fmt.Println("Erro criado cliente do kafka")
		panic(err)
	}
	stageChan = make(chan int)
}

func setup() {
	var t, e = topic.LoadTopicConfig("topics.json")
	if e != nil {
		fmt.Println("Erro ao importar arquivo json de topicos")
		panic(e)
	}
	createTopics(t)
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

// func createTopic(v topic.Topics) {
// 	kafkaAdmClient.DeleteTopics(context.Background(), v.Name)
// }

// não precisaria criar um go routine, quem chama ela poderia chamar usando uma rotina
func newStageJob(begin, end int, f func(stopChan chan bool)) {
	stopChan := make(chan bool)
	go func() { // inicia a rotina que vai observar o job fechando o canal quando for hora de parar
		for {
			select {
			case <-stopChan: // a rotina parou, como parar avisar para a main que o estágio ?
				close(stopChan)
				return
			case stage := <-stageChan:
				if stage == begin { // verificar se é possivel matar a rotina sem usar o canal
					go f(stopChan) // inicia a rotina
				} else if stage == end { //recebei de fora a info que o job tem que parar
					close(stopChan)
					return
				}
			}
		}
	}()
}
// go routine // stage 1 ()  
// variavelStage = 2
// stageChan <- variavelStage


// stage 1

// stage 2

// fazer um arrray de WG para que cada job  seja inserido no estagio que ele vai terminar
func main() {
	var e error

	defer teardown()
	internalSetup()
	setup()

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
		fmt.Printf("e: %v\n", e)
	} else {
		fmt.Printf("c: %v\n", container)

		cerr := client.StartContainer(container.ID, container.HostConfig)

		// olhar healthcheck

		if cerr != nil {
			fmt.Printf("cerr: %v\n", cerr)
			return
		}
		time.Sleep(time.Second * 20)

		fmt.Print("parando container: %v\n", container.ID)
		cerr = client.StopContainer(container.ID, 5)
		if cerr != nil {
			fmt.Print("could not stop contaier: %v\n", cerr)
		}

	}
}

func remove(client *docker.Client, container *docker.Container) {
	if err := client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID}); err != nil {
		fmt.Printf("err: %v\n", err)
	}
}

func teardown() {
	if container != nil {
		fmt.Printf("removendo container: %v\n", container.ID)
		remove(client, container)
	} else {
		fmt.Printf("container nulo: %v\n", container.ID)
	}
	if kafkaAdmClient != nil {
		kafkaAdmClient.Close()
	}
	if kafkaProducer != nil {
		kafkaProducer.Close()
	}
}

// WG.add(t)

// <-WG

// asdada
