package main

import (
	"fmt"
	"log"
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

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func main() {
	args := os.Args
	configFile := "config.json"
	if len(args) > 1 {
		if fileExists(args[1]) {
			configFile = args[1] // config file will be the first arg
		} else {
			log.Printf("Could not use the received config: %s. Will use the default config: %s\n", args[1], configFile)
		}
	}
	if fileExists(configFile) {
		log.Printf("Using config: %s.\n", configFile)
	} else {
		panic("Could not find a config file.")
	}
	conf, err := config.GetConfig(configFile)
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
