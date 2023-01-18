package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"teste.com/docker/dockertest"
	"teste.com/docker/kafkatest"
)

func waitForHealthStatus(cli *client.Client, id string, status string, interval, tryies int) (c chan (string), err error) {
	if id == "" {
		return nil, errors.New("empty id")
	}

	channel := make(chan (string))
	go func() {
		var lastStatus string
		for i := 0; i < tryies; i++ {
			s2, _ := cli.ContainerInspect(context.Background(), id)

			if s2.State.Health != nil {
				lastStatus = s2.State.Health.Status
				if s2.State.Health.Status == status {
					fmt.Printf("[%s, %s] %s....\n", id, s2.Name, status)
					channel <- status
					return
				} else {
					fmt.Printf("[%s, %s] NOT %s....\n", id, s2.Name, status)
				}
			} else {
				println(s2.State.Health)
				fmt.Printf("[%s, %s]No Health status....\n", id, s2.Name)
				continue
			}
			time.Sleep(time.Millisecond * time.Duration(interval))
		}
		channel <- lastStatus
	}()

	return channel, nil
}

func main() {
	// dockerTests()
	kafkaApiTests()
}

func kafkaApiTests() {
	// kafkatest.ListClusters()
	// kafkatest.ListTopics()
	kafkatest.GetTopic("demanda-submetida")
	// kafkatest.ListConsumerGroups()
}

func dockerTests() {
	containerConfig, err := dockertest.LoadDockerConfig("sample-container-config.json")
	if err != nil {
		panic(err)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	containers, err := runOdjContainers(1, containerConfig, cli)

	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	containerHealth := make(map[string]string, len(containers))
	for _, container := range containers {
		s1, e := cli.ContainerInspect(context.Background(), container)
		if e != nil {
			fmt.Printf("Erro ao fazer inspect do container [%s]\n", container)
			fmt.Println(e)
			continue
		}

		if s1.State.Health != nil {
			wg.Add(1)
			go func(id string) {
				fmt.Printf("[%s]will check for health\n", id)
				c, err := waitForHealthStatus(cli, id, types.Healthy, 2000, 20)
				if err != nil {
					fmt.Printf("Erro ao checar health do container %s\n", id)
				}
				r := <-c
				containerHealth[container] = r
				if r != "" && r == types.Healthy {
					fmt.Printf("[%s] Healthcheck ficou health (%s)\n", id, r)
				} else {
					fmt.Printf("[%s] Healthcheck did not become %s, lastHealtheck: %s\n", id, types.Healthy, r)
				}
				wg.Done()
			}(container)
		} else {
			fmt.Printf("****%s Health status nil (no healthcheck?\n", container[:10])
			continue
		}
	}
	wg.Wait()

	// TESTE: 1 container:
	// c := containers[0]
	// if containerHealth[c] == types.Healthy {
	// 	fmt.Printf("Iniciando teste com container %s\n", c)

	// }
	// if cWithHealthId != "" {
	// 	cli.ContainerWait(context.Background(), cWithHealthId, container.WaitConditionNextExit)
	// 	cli.ContainerInspect(context.Background(), cWithHealthId)
	// 	reader, err := cli.ContainerStats(context.Background(), cWithHealthId, true)
	// 	if err != nil {
	// 		panic("Erro pegando stats")
	// 	}
	// 	bPointer := make([]byte, 100)

	// 	for {
	// 		n, err := reader.Body.Read(bPointer)
	// 		if err != nil || n == 0 {
	// 			break
	// 		}
	// 		fmt.Println(string(bPointer))
	// 	}

	// }
}

func runOdjContainers(n int, c dockertest.DockerConfigFile, client *client.Client) ([]string, error) {
	cts := make([]string, n)
	r, e := client.ContainerCreate(context.Background(), &container.Config{
		Image: c.Config.Image,
		Env:   []string{"ODR_KAFKA_HOST=kafka:9092", "ODR_TJRJ_PROCESSOS_URI=http://mock-api-tjrj-processos:3001"},
		Healthcheck: &container.HealthConfig{
			Test:        []string{"CMD", "curl", "-f", "http://localhost:8080/q/health/live"},
			Retries:     10,
			Timeout:     time.Second,
			StartPeriod: time.Second,
		}},
		&container.HostConfig{Binds: c.HostConfig.Binds}, &network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{"dev-dockercompose_default": {}}}, nil, "teste_teste123")
	if e != nil {
		panic(e)
	}
	e = client.ContainerStart(context.Background(), r.ID, types.ContainerStartOptions{})
	if e != nil {
		panic(e)
	}
	cts = append(cts, r.ID)
	return cts, nil
	// for i := 0; i < n; i++ {
	// 	containerName := "monitor_teste_" + fmt.Sprint(i)
	// 	dockerCreatorConfig := client.CreateContainerOptions{
	// 		Name:             containerName,
	// 		Config:           &c.Config,     // Contém a imagem a ser executada
	// 		HostConfig:       &c.HostConfig, // Contém os binds de arquivos
	// 		NetworkingConfig: &c.NetworkingConfig,
	// 	}
	// 	container, e := client.CreateContainer(dockerCreatorConfig) // TODO: Mover para a chamada docker
	// 	if e != nil {
	// 		fmt.Printf("e: %v\n", e)
	// 		return cts, e
	// 	}
	// 	cerr := client.StartContainer(container.ID, container.HostConfig)
	// 	if cerr != nil {
	// 		fmt.Printf("cerr: %v\n", cerr)
	// 		return cts, cerr
	// 	}
	// 	cts = append(cts, container)

	// }

}
