package producer

import (
	"reflect"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var mockConfig = &kafka.ConfigMap{"test.mock.num.brokers": 1}

func createMockProducer(t *testing.T, c *kafka.ConfigMap) *kafka.Producer {
	mockProducer, err := kafka.NewProducer(c)
	if err != nil {
		if err != nil {
			t.Errorf(" %s: *Precondition error*. Could not create mockProducer error = %v", t.Name(), err)
		}
	}
	return mockProducer
}

func Test_produce(t *testing.T) {
	mockProducer := createMockProducer(t, mockConfig)
	emptyMsg := &kafka.Message{}
	topic := "topic"
	validMsg := &kafka.Message{Value: []byte{123}, Key: nil, TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}}
	type args struct {
		p            *kafka.Producer
		msg          *kafka.Message
		deliveryChan chan kafka.Event
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Invalid message", wantErr: true, args: args{p: mockProducer, msg: emptyMsg}},
		{name: "valid message", wantErr: false, args: args{p: mockProducer, msg: validMsg}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := produce(tt.args.p, tt.args.msg, tt.args.deliveryChan); (err != nil) != tt.wantErr {
				t.Errorf("produce() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_producer_Write(t *testing.T) {
	mockP := producer{
		kProducer: createMockProducer(t, mockConfig),
		pChannel:  make(chan kafka.Event),
	}

	type args struct {
		b     []byte
		topic string
		key   string
	}
	tests := []struct {
		name string
		p    *producer
		args args
	}{
		{name: "empty topic", p: &mockP, args: args{b: []byte{3}, topic: "", key: "chave"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go tt.p.Write(tt.args.b, tt.args.topic, tt.args.key)
			m := <-tt.p.pChannel
			_, ok := m.(kafka.Error)
			if !ok {
				t.Errorf("Write() msg got from channel is not an error. Got: %+v , expected: Kafka.Error \n", m)
			}
		})
	}
}

func Test_createProducer(t *testing.T) {
	mockProducer := createMockProducer(t, mockConfig)

	type args struct {
		c *kafka.ConfigMap
	}
	tests := []struct {
		name    string
		args    args
		want    *kafka.Producer
		wantErr bool
	}{
		{name: "Simple mock argument", want: mockProducer, args: args{mockConfig}},
		{name: "Empty kafka config", want: mockProducer, args: args{&kafka.ConfigMap{}}},
		{name: "Wrong kafka config", want: nil, wantErr: true, args: args{&kafka.ConfigMap{"invalid.key": "error"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := createProducer(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("createProducer() error = %v, wantErr %v, producer: %#v", err, tt.wantErr, got)
				return
			}
			if !(reflect.TypeOf(got) == reflect.TypeOf(tt.want)) {
				t.Errorf("createProducer() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_buildProducer(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := buildProducer(); (err != nil) != tt.wantErr {
				t.Errorf("buildProducer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetProducer(t *testing.T) {
	tests := []struct {
		name string
		want Producer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetProducer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProducer() = %v, want %v", got, tt.want)
			}
		})
	}
}
