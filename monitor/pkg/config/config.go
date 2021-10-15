package config

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	KafkaUrl, KafkaStartOffset, MonitorTopic string
}

func LoadConfig(jsonFile string) (*Config, error) {
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
