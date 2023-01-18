package kafkatest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var clusterId string
var client = http.Client{}

func ListClusters() ([]string, error) {
	url := "http://localhost:8082/v3/clusters"
	get(url)
	return nil, nil
}

func ListTopics() {
	if clusterId == "" {
		setClusterID()
	}
	url := "http://localhost:8082/v3/clusters/" + clusterId + "/topics"
	resp, err := get(url)
	if err != nil {
		return
	}
	fmt.Printf("****Teste, resp: %s\n", jsonPrettyPrint(string(resp)))
}

func GetTopic(name string) {
	if clusterId == "" {
		setClusterID()
	}
	url := "http://localhost:8082/v3/clusters/" + clusterId + "/topics/" + name
	resp, err := get(url)
	if err != nil {
		return
	}
	fmt.Printf("****Teste, resp: %s\n", jsonPrettyPrint(string(resp)))
}

func ListConsumerGroups() {
	if clusterId == "" {
		setClusterID()
	}
	url := "http://localhost:8082/v3/clusters/" + clusterId + "/consumer-groups"
	resp, err := get(url)
	if err != nil {
		return
	}
	fmt.Printf("****Consumer groups, resp: %s\n", jsonPrettyPrint(string(resp)))
}

func setClusterID() error {
	url := "http://localhost:8082/v3/clusters"
	resp, err := get(url)
	if err != nil {
		fmt.Printf("Could not get cluster id...\n%s\n", err)
		return err
	}
	// fmt.Printf("#### %s\n", string(resp))

	var iCluster GenericKafkaResponse
	// var any map[string]interface{}
	err = json.Unmarshal(resp, &iCluster)
	if err != nil {
		fmt.Printf("Could not unmarshal cluster id...\n%s\n", err)
		return err
	} else if iCluster.Data[0]["cluster_id"] == nil || iCluster.Data[0]["cluster_id"] == 0 {
		fmt.Printf("Cluster id not found...\n%s\n", iCluster.Data)
		return err
	}
	id, ok := iCluster.Data[0]["cluster_id"].(string)
	if !ok {
		fmt.Printf("cluster id not a string...\n%s\n", err)
		return errors.New("cluster id not string")
	}
	clusterId = id
	return nil
}

func get(url string) ([]byte, error) {
	emptyBody := "{}"
	bufferBody := bytes.NewBuffer([]byte(emptyBody))
	req, err := http.NewRequest(http.MethodGet, url, bufferBody)
	if err != nil {
		return nil, err
	}
	fmt.Printf("-->Call %s\n", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("-->Error calling %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	bodyResp := bufio.NewScanner(resp.Body)
	response := make([]byte, 0)
	fmt.Printf("<--Response %s\n", resp.Status)
	for bodyResp.Scan() {
		response = append(response, bodyResp.Bytes()...)
	}
	return response, nil
}

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}
