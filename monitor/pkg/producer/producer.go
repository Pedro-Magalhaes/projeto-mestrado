/*
	Pacote que implementa um singleton para que os outros pacotes possam produzir mensagens para o kafka

	Autor: Pedro Magalhães
*/
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

/*
	Função interna que vai produzir para o producer do kafka a mensagem recebida por parametro
*/
func produce(p *kafka.Producer, msg *kafka.Message, deliveryChan chan kafka.Event) error {
	return p.Produce(msg, deliveryChan)
}

/*
	Função publica para fechar o produtor
*/
func (p *producer) Close() {
	p.kProducer.Close()
}

/*
	Função publica que trata a escrita do producer
	Recebe os bytes do conteudo da mensagem (b), o tópico que deve ser enviada a mensagem e a chave do job
	tópico
	Se houver erro na produção o erro será enviado para o canal do produtor
*/
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

/*
	Função que faz a escrita para uma partição especifica de um tópico.
	Assume que os parametros são corretos e qualquer erro vai ser apenas logado no console
*/
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

/*
	Cria um novo produtor do kafka com as configurações recebidas no kafka.ConfigMap
*/
func createProducer(c *kafka.ConfigMap) (*kafka.Producer, error) {
	return kafka.NewProducer(c)
}

/*
	Função que faz a inicialização do producer e dá start na rotina que vai ficar ouvindo o canal
	do produtor para logar algum erro de envio de mensagens
*/
func buildProducer() error {
	var err error
	conf, err = config.GetConfig("config.json")
	if err != nil {
		return err
	}
	prod, err := createProducer(&kafka.ConfigMap{"bootstrap.servers": conf.KafkaUrl, "client.id": time.Now().GoString(),
		"acks": "all"})
	if err != nil {
		return err
	}
	channel := make(chan kafka.Event, 1000)

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

/*
	Função que vai ser usada para obter o producer nos outros módulos.
	Retorna a instância do produtor, se ela for nula chama a função buildProducer para inicializar o producer
*/
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
