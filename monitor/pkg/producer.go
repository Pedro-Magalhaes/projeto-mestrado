package coord

import (
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer interface {
	Write([]byte)
}

type producer struct {
	kProducer kafka.Producer
	pChannel  chan kafka.Event
}

func (p *producer) Write(b []byte) {
	p.kProducer.Produce(&kafka.Message{Value: b}, p.pChannel)
}

var instance Producer

var once sync.Once

func buildProducer() {
	instance = &producer{
		// kProducer: kafka.NewProducer(),
	}
}

func GetProducer() *Producer {
	once.Do(buildProducer)
	return &instance
}
