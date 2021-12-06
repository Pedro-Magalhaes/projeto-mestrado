package config

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type Config struct {
	KafkaUrl               string `json:"kafkaUrl"`
	KafkaStartOffset       string `json:"kafkaStartOffset"`
	MonitorTopic           string `json:"monitorTopic"`
	JobInfoTopic           string `json:"jobInfoTopic"`
	StateTopic             string `json:"stateTopic"`
	BasePath               string `json:"basePath"`
	ChunkSize              uint   `json:"chunkSize"`
	ConsumerTimeout        int    `json:"consumerTimeout"`
	ConsumerSessionTimeout int    `json:"consumerSessionTimeout"`
}

var c *Config

var defaultJsonFile = "config.json"

func loadConfig(jsonFile string) (*Config, error) {
	file, err := os.Open(jsonFile)
	if err != nil {
		return nil, err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	conf := Config{}

	err = json.Unmarshal(bytes, &conf)

	if err != nil {
		return nil, err
	}
	return &conf, nil
}

// return the global config the jsonFile param
// is used only on the first time its called if
// a empty string is used the defaul value will be
// used
func GetConfig(jsonFile string) (Config, error) {
	if c != nil {
		return *c, nil
	}
	if jsonFile == "" {
		jsonFile = defaultJsonFile
	}
	var err error
	c, err = loadConfig(jsonFile)
	if err != nil {
		log.Println("Error loading config file: ", jsonFile)
		panic(err)
	}
	return *c, nil
}
