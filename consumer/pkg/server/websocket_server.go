package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/websocket"
	"github.com/pfsmagalhaes/consumer/pkg/config"
)

var fourSecondsInNano = time.Duration(4000000000)

type producerMsg struct {
	Path    string `json:"path"`
	Watch   bool   `json:"watch"`
	Project string `json:"project"`
}

type client struct {
	currOffset int64
	topic      string
}

type FileChunkMsg struct {
	Msg    string `json:"msg"`
	Offset int64  `json:"offset"`
	Lenth  int    `json:"lenth"`
}

var producer *kafka.Producer
var c *config.Config

func Start() {
	c, _ = config.LoadConfig("config.json")
	if err := buildProducer(); err != nil {
		log.Panicln("Could not create producer! exiting. Error: ", err)
	}
	start(c.ServerPort)
}

func buildProducer() error {
	var err error
	producer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": c.KafkaUrl, "client.id": time.Now().GoString(),
		"acks": "all"})
	if err != nil {
		return err
	}
	return nil
}

func newKafkaConsumer(path string, ws *websocket.Conn) error {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{"bootstrap.servers": c.KafkaUrl,
		"client.id": time.Now().GoString(),
		"group.id":  time.Now().GoString()})
	if err != nil {
		log.Println("Error creating consumer: ", err)
		return err
	}
	path = "test_files/" + path
	topic := strings.ReplaceAll(path, "/", "__")
	err = c.Subscribe(topic, nil) // will not care for rebalance
	if err != nil {
		log.Printf("Error, could not connect to topic %s. Error: %s\n", topic, err)
		for {
			time.Sleep(fourSecondsInNano) // 4 segundos
			log.Printf("Trying to reconect to topic %s.\n", topic)
			err = c.Subscribe(topic, nil) // will not care for rebalance
			if err == nil {
				break
			}
		}
	}
	log.Printf("conected to topic %s.\n", topic)
	closedConn := make(chan bool)
	// TODO: não funciona! preciso ler durante o loop e pegar o "erro" de conexão fechada
	ws.SetCloseHandler(func(code int, text string) error {
		log.Printf("connection closed %s.\n", topic)
		closedConn <- true
		return nil
	})
	go func(consumer *kafka.Consumer) {
		defer ws.Close()
		state := &client{currOffset: 0, topic: topic}
		for {
			select {
			case <-closedConn:
				log.Printf("Stopping consumer routine\n")
				return
			default:
				ev := consumer.Poll(2000)
				switch e := ev.(type) {
				case *kafka.Message:
					if strings.Compare(*e.TopicPartition.Topic, topic) == 0 {
						msg := FileChunkMsg{}
						if err := json.Unmarshal(e.Value, &msg); err != nil {
							fmt.Println("Error msg does not respect msg interface!")
							return
						}
						if msg.Offset < state.currOffset {
							log.Printf("Offset smaller than current. Got %d, Received: %d\n", msg.Offset, state.currOffset)
							continue
						} else if msg.Offset > state.currOffset {
							log.Printf("Offset bigger than current. Got %d, Received: %d\n", msg.Offset, state.currOffset)
						}
						state.currOffset = msg.Offset + int64(msg.Lenth)
						if err := ws.WriteJSON(msg); err != nil {
							log.Println("ERROR Writing msg to websocket! Stoping routine.", err)
							ws.Close()
							return
						}
					} else {
						log.Println("Unknown topic: %w", e.TopicPartition)
					}
				case kafka.PartitionEOF:
					log.Printf("%% Reached %v\n", e)
				case kafka.Error:
					log.Printf("%% CONSUMER ERROR %v\n", e)
					log.Println(e.Code())
					log.Println(e.IsRetriable())
					log.Println(e.IsFatal())
					if e.IsFatal() {
						return
					}
				default: // poll deu timeout.
					// fmt.Printf("Ignored %v\n", e)
				}
			}
		}
	}(c)
	return nil
}

func sendMsg(msg producerMsg, jobid string) error {
	value, e := json.Marshal(msg)
	if e != nil {
		log.Default().Println("Error could not marshal msg.", e)
		return e
	}
	delivery_chan := make(chan kafka.Event, 10000)
	defer close(delivery_chan)
	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &c.ProducerTopic, Partition: kafka.PartitionAny},
		Value:          []byte(value),
		Key:            []byte(jobid)},
		delivery_chan)

	if err != nil {
		return err
	}
	ev := <-delivery_chan
	m := ev.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
		return m.TopicPartition.Error
	} else {
		fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}
	return nil
}

func buildWS(ws *websocket.Conn, msg producerMsg, jobid string) {

	err := sendMsg(msg, jobid)
	if err != nil {
		log.Default().Println("Error could not produce msg.", err)
		return
	}
	log.Println("Creating consumer with path: ", msg.Path)
	newKafkaConsumer(msg.Path, ws)
}

var upgrader = websocket.Upgrader{}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	params := r.URL.Query()
	log.Println(params)
	jobid := params.Get("jobid")

	if jobid == "" {
		log.Default().Printf("Error jobid not received. Closing connection to: %s", r.RemoteAddr)
		ws.Close()
	}
	projectid := params.Get("projectid")
	if projectid == "" {
		log.Default().Printf("Error projectid not received. Closing connection to: %s", r.RemoteAddr)
		ws.Close()
	}
	path := params.Get("path")
	if path == "" {
		log.Default().Printf("Error path not received. Closing connection to: %s", r.RemoteAddr)
		ws.Close()
	}
	go buildWS(ws, producerMsg{Path: path, Watch: true, Project: projectid}, jobid)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.FileServer(http.Dir("html")).ServeHTTP(w, r)
}

func completeJob(w http.ResponseWriter, r *http.Request) {
	jobid := r.Form.Get("jobid")
	if jobid == "" {
		log.Default().Printf("Error path not received. Closing connection to: %s", r.RemoteAddr)
		http.Error(w, "jobid missing", 400)
		return
	}
	delivery_chan := make(chan kafka.Event, 1000)
	producer.Produce(&kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &c.JobTopic, Partition: kafka.PartitionAny}, Value: []byte("{'status': 'finished'}"), Key: []byte(jobid)}, delivery_chan)
}

func start(port string) {
	http.Handle("/", http.FileServer(http.Dir("./html")))
	http.HandleFunc("/complete", completeJob)
	http.HandleFunc("/ws", serveWs)
	addr := strings.Join([]string{"localhost", port}, ":")
	log.Fatal(http.ListenAndServe(addr, nil))
}
