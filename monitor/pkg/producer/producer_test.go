package producer

import (
	"reflect"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func Test_produce(t *testing.T) {
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.Write(tt.args.b, tt.args.topic, tt.args.key)
		})
	}
}

func Test_createProducer(t *testing.T) {
	type args struct {
		c *kafka.ConfigMap
	}
	tests := []struct {
		name    string
		args    args
		want    *kafka.Producer
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createProducer(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("createProducer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createProducer() = %v, want %v", got, tt.want)
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
