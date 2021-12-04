package producer

import (
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pfsmagalhaes/monitor/pkg/config"
)

type Producer interface {
	Write([]byte, string, string)
	WriteToPartition(value []byte, key []byte, topic string, partition int32)
	Close()
}

type producer struct {
	kProducer *kafka.Producer
	pChannel  chan kafka.Event
}

var instance Producer

var conf config.Config

func produce(p *kafka.Producer, msg *kafka.Message, deliveryChan chan kafka.Event) error {
	return p.Produce(msg, deliveryChan)
}

func (p *producer) Close() {
	p.kProducer.Close()
}

func (p *producer) Write(b []byte, topic string, key string) {
	var topicKey []byte
	if key == "" {
		topicKey = nil
	} else {
		topicKey = []byte(key)
	}
	err := produce(p.kProducer, &kafka.Message{Value: b, Key: topicKey, TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}}, p.pChannel)
	if err != nil {
		fmt.Println("ERROR: erro escrevendo para o kakfa. Topic:", topic)
		fmt.Println(err.Error())
		e := kafka.NewError(0, "Error", true)
		p.pChannel <- e
	}
}

func (p *producer) WriteToPartition(value []byte, key []byte, topic string, partition int32) {
	if partition < 0 {
		partition = kafka.PartitionAny
	}
	err := produce(p.kProducer, &kafka.Message{Value: value, Key: key, TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: partition}}, p.pChannel)
	if err != nil {
		fmt.Println("ERROR: erro escrevendo para o kakfa. Topic:", topic)
		fmt.Println(err.Error())
	}
}

func createProducer(c *kafka.ConfigMap) (*kafka.Producer, error) {
	return kafka.NewProducer(c)
}

func buildProducer() error {
	var err error
	conf, err = config.GetConfig("config.json") // TODO: melhorar config
	if err != nil {
		return err
	}
	prod, err := createProducer(&kafka.ConfigMap{"bootstrap.servers": conf.KafkaUrl, "client.id": time.Now().GoString(),
		"acks": "all"})
	if err != nil {
		return err
	}
	channel := make(chan kafka.Event, 1000) // TODO: checar numero

	instance = &producer{
		kProducer: prod,
		pChannel:  channel,
	}
	go func() {
		for {
			e := <-channel
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				}
			}
		}
	}()
	return nil
}

func GetProducer() Producer {
	if instance == nil {
		log.Println("Creating producer")
		err := buildProducer()
		if err != nil {
			return nil
		}
	}
	return instance
}
