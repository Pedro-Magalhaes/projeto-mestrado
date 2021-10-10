package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	coord "github.com/pfsmagalhaes/monitor/pkg"
)

func captureInterrupt(channel chan bool) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGTERM,
		syscall.SIGINT)
	go func() {
		<-sigc
		fmt.Println("ctrl+c pressed")
		channel <- true
	}()
}

func waitTermination(normalChan, abortChan chan bool) {
	select {
	case <-normalChan:
		fmt.Println("Stoping main, workers have finished")
		break
	case <-abortChan:
		fmt.Println("Main Aborting")
		break
	}
}

func main() {
	monitorState := coord.SafeMap{V: make(map[string]bool)}
	end := make(chan bool)   // used to receive termination notice from the coordinator
	abort := make(chan bool) // used to receive the interrupt signal
	function, err := coord.Start(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest"},
		&monitorState)
	if err != nil {
		fmt.Print("Bye World!")
		fmt.Print(err)
		os.Exit(1)
	}

	go function(end)

	captureInterrupt(abort)
	waitTermination(end, abort)
}
