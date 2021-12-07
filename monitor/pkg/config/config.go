/*
	Pacote que vai carregar a configuração do monitor via arquivo de configuração

	Autor: Pedro Magalhães
*/
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

var DefaultJsonFile = "config.json"

// Funçao que abre o arquivo json e faz o parse para o objeto do tipo Config
// é usada na função GetConfig quando a configuração ainda não foi carregada
// Falha se houver erro ao abrir e ler o arquivo ou se o o arquivo não conter um objeto que
// possa ser mapeado para o objeto Config
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

/*
	Função que é usada para pegar a configuração de um arquivo json.
	Recebe uma string que indica a localização da configuração, se receber
	a string vazia ( "" ), utiliza o valor default DefaultJsonFile. Essa
	função só vai rodar uma vez, nas chamadas subsequentes vai retornar a configuração
	já obtida em uma chamada anterior, independente do valor da string recebida
	Retorna a configuração e um erro. Em caso de sucesso a configuração é
	diferente de nulo e o erro nulo. Em caso de erro é o oposto.
*/
func GetConfig(jsonFile string) (Config, error) {
	if c != nil {
		return *c, nil
	}
	if jsonFile == "" {
		jsonFile = DefaultJsonFile
	}
	var err error
	c, err = loadConfig(jsonFile)
	if err != nil {
		log.Println("Error loading config file: ", jsonFile)
		panic(err)
	}
	return *c, nil
}
