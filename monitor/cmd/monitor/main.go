package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	coord "github.com/pfsmagalhaes/monitor/pkg"
	"github.com/pfsmagalhaes/monitor/pkg/config"
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
	conf, err := config.GetConfig("config.json")
	group := "myGroup"
	server := "localhost:9092"
	offset := "earliest"
	if err != nil {
		fmt.Println("ERROR LOADING CONFIG FILE...")
		fmt.Println(err.Error())
	} else {
		fmt.Println("CONFIG LOADED")
		server = conf.KafkaUrl
		offset = conf.KafkaStartOffset
	}

	endChannel := make(chan bool)   // used to receive termination notice from the coordinator
	abortChannel := make(chan bool) // used to receive the interrupt signal

	coordinatorThread, err := coord.Create(&kafka.ConfigMap{
		"bootstrap.servers":        server,
		"group.id":                 group,
		"allow.auto.create.topics": true,
		"auto.offset.reset":        offset},
		conf)
	if err != nil {
		fmt.Print("Bye World!")
		fmt.Print(err)
		os.Exit(1)
	}

	go coordinatorThread(endChannel)

	captureInterrupt(abortChannel)
	waitTermination(endChannel, abortChannel)
}
