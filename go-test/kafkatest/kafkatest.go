package kafkatest

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Kafka struct {
	server        string
	kafkaProducer *kafka.Producer
	// kafkaConsumer *kafka.Consumer
	kafkaAdmClient *kafka.AdminClient
	// Colocar array para mensagens de cada tópico? O que ocorreria em um topico muito movimentado
}

func newKafka(server string) *Kafka {
	k := Kafka{
		server: server,
	}
	var err error
	kConfig := kafka.ConfigMap{"bootstrap.servers": k.server, "acks": "all"}
	k.kafkaAdmClient, err = kafka.NewAdminClient(&kConfig)
	if err != nil {
		fmt.Println("Erro criando adm do kafka")
		panic(err)
	}

	k.kafkaProducer, err = kafka.NewProducer(&kConfig)
	if err != nil {
		fmt.Println("Erro criando producer do kafka")
		panic(err)
	}
	return &k
}
