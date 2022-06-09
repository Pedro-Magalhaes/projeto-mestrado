// package main

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"time"

// 	docker "github.com/fsouza/go-dockerclient"
// )

// func execInContainer(command string) {
// 	endpoint := os.Getenv("DOCKER_HOST")
// 	if endpoint == "" {
// 		endpoint = "unix:///var/run/docker.sock"
// 	}
// 	client, err := docker.NewClient(endpoint)
// 	if err != nil {
// 		fmt.Print("err.Error(): %v\n", err.Error())
// 		return
// 	}
// 	fmt.Print("Created client")

// 	//Pull image from Registry, if not present
// 	imageName := "golang:1.15"

// 	//Try to create a container from the imageID
// 	config := docker.Config{
// 		Image:        imageName,
// 		OpenStdin:    true,
// 		StdinOnce:    true,
// 		AttachStdin:  true,
// 		AttachStdout: true,
// 		AttachStderr: true,
// 		Tty:          true,
// 	}

// 	opts2 := docker.CreateContainerOptions{Config: &config}
// 	container, err := client.CreateContainer(opts2)
// 	if err != nil {
// 		fmt.Print("err.Error(): %v\n", err.Error())
// 		return
// 	} else {
// 		defer func() {
// 			err = client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID})
// 			if err != nil {
// 				fmt.Print("err.Error(): %v\n", err.Error())
// 				return
// 			}
// 			fmt.Print("Removed container with ID", container.ID)
// 		}()
// 	}
// 	fmt.Print("Created container with ID", container.ID)

// 	//Try to start the container
// 	err = client.StartContainer(container.ID, &docker.HostConfig{})
// 	if err != nil {
// 		fmt.Print("err.Error(): %v\n", err.Error())
// 		return
// 	} else {
// 		// And once it is done with all the commands, remove the container.
// 		defer func() {
// 			err = client.StopContainer(container.ID, 0)
// 			if err != nil {
// 				fmt.Print("err.Error(): %v\n", err.Error())
// 				return
// 			}
// 			fmt.Print("Stopped container with ID", container.ID)
// 		}()
// 	}
// 	fmt.Print("Started container with ID", container.ID)

// 	// var (
// 	// 	dExec *docker.Exec
// 	// )
// 	// cmd := []string{"bash", "-c", command}

// 	// de := docker.CreateExecOptions{
// 	// 	AttachStderr: true,
// 	// 	AttachStdin:  true,
// 	// 	AttachStdout: true,
// 	// 	Tty:          true,
// 	// 	Cmd:          cmd,
// 	// 	Container:    container.ID,
// 	// }
// 	// fmt.Print("CreateExec")
// 	// if dExec, err = client.CreateExec(de); err != nil {
// 	// 	fmt.Print("err.Error(): %v\n", err.Error())
// 	// 	return
// 	// }
// 	// fmt.Print("Created Exec")
// 	// execId := dExec.ID
// 	time.Sleep(time.Second * 20)
// 	// Context with timeout!!!
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
// 	fmt.Print("ctx: %v\n", ctx)
// 	defer cancel()

// 	// opts := docker.StartExecOptions{
// 	// 	OutputStream: os.Stdout,
// 	// 	ErrorStream:  os.Stderr,
// 	// 	InputStream:  os.Stdin,
// 	// 	RawTerminal:  true,
// 	// 	Context:      ctx,
// 	// }

// 	// fmt.Print("StartExec")
// 	// if err = client.StartExec(execId, opts); err != nil {
// 	// 	fmt.Print("StartExec Error: %s", err)
// 	// 	return
// 	// }

// }

// func main() {

// 	execInContainer("sleep 60")
// }
