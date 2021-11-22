package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         "Testes_produção",
		"acks":              "all"})

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Digite o tópico e a mensagem desejada da segunte forma <topico>;<mensagem>")
		fmt.Print("-> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Erro ao ler entrada", err)
			os.Exit(1)
		}
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		splitText := strings.Split(text, ";")

		if len(splitText) < 2 {
			fmt.Println("o tópico e a mensagem tem que ser separados por ';' o texto digitado foi:")
			fmt.Println(text)
			continue
		}

		topic := strings.TrimSpace(splitText[0])
		value := strings.Join(splitText[1:], " ")
		fmt.Println("enviando mensagem: ", value)

		// delivery_chan := make(chan kafka.Event, 10000)
		p.ProduceChannel() <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(value)}

		// err = p.Produce(&kafka.Message{
		// 	TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		// 	Value:          []byte(value)},
		// 	delivery_chan,
		// )

		e := <-p.Events()
		m := e.(*kafka.Message)
		fmt.Println(m)
	}
}
