package producer

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pfsmagalhaes/monitor/pkg/config"
)

type Producer interface {
	Write([]byte, string)
}

type producer struct {
	kProducer *kafka.Producer
	pChannel  chan kafka.Event
}

var instance Producer

var conf *config.Config

func (p *producer) Write(b []byte, topic string) {
	err := p.kProducer.Produce(&kafka.Message{Value: b, TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}}, p.pChannel)
	if err != nil {
		fmt.Println("ERROR: erro escrevendo para o kakfa. Topic:", topic)
		fmt.Println(err.Error())
	}
	p.kProducer.Flush(10) // TODO: Melhorar lógica de produção, deixar o producer enviar no passo dele?
}

func buildProducer() error {
	var err error
	conf, err = config.LoadConfig("config.json") // TODO: melhorar config
	if err != nil {
		return err
	}
	prod, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": conf.KafkaUrl, "client.id": time.Now().GoString(),
		"acks": "all"})
	if err != nil {
		return err
	}
	channel := make(chan kafka.Event, 100) // TODO: checar numero

	instance = &producer{
		kProducer: prod,
		pChannel:  channel,
	}
	return nil
}

func GetProducer() Producer {
	if instance == nil {
		err := buildProducer()
		if err != nil {
			return nil
		}
	}
	return instance
}
