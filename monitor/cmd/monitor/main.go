package main

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	coord "github.com/pfsmagalhaes/monitor/pkg"
)

func main() {
	// monitoring := coord.SafeMap{v: make(map[string]bool)}
	monitorState := coord.SafeMap{V: make(map[string]bool)}
	end := make(chan bool)
	function, err := coord.Start(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest"},
		&monitorState)
	if err != nil {
		fmt.Print("Bye World!")
		fmt.Print(err)
	} else {
		go function(end)
	}
	<-end
}
