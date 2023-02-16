package main

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
)

func main() {
    ctx := context.Background()

    requests := testcontainers.ParallelContainerRequest{
        {
            ContainerRequest: testcontainers.ContainerRequest{

                Image: "nginx",
                ExposedPorts: []string{
                    "10080/tcp",
                },
            },
            Started: true,
        },
        {
            ContainerRequest: testcontainers.ContainerRequest{

                Image: "nginx",
                ExposedPorts: []string{
                    "10081/tcp",
                },
            },
            Started: true,
        },
    }

    res, err := testcontainers.ParallelContainers(ctx, requests, testcontainers.ParallelContainersOptions{})
    if err != nil {
        e, ok := err.(testcontainers.ParallelContainersError)
        if !ok {
            log.Fatalf("unknown error: %v", err)
        }

        for _, pe := range e.Errors {
            fmt.Println(pe.Request, pe.Error)
        }
        return
    }

    for _, c := range res {
        c := c
        defer func() {
            if err := c.Terminate(ctx); err != nil {
                log.Fatalf("failed to terminate container: %s", c)
            }
        }()
    }
}
