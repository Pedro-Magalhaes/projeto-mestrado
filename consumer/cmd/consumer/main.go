package main

import (
	"log"

	"github.com/pfsmagalhaes/consumer/pkg/server"
)

func main() {
	log.Println("Starting main")
	server.Start()
}
