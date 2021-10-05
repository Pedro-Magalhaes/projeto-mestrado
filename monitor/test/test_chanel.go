package main

import (
	"fmt"
	"sync"
	"time"
)

// import (

// )

type loopControl struct {
	mu          sync.Mutex
	keepWorking bool
}

func main() {
	// watcher, err := fsnotify.NewWatcher()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer watcher.Close()
	control := loopControl{}
	control.keepWorking = true
	done := make(chan bool)
	go func() {
		for {
			control.mu.Lock()
			if !control.keepWorking {
				control.mu.Unlock()
				done <- true
				break
			}
			control.mu.Unlock()
			time.Sleep(300000000) //nano
			fmt.Printf(".")
		}
	}()
	// err = watcher.Add("./temp/a.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	for i := 0; i < 10; i++ {
		time.Sleep(300000000)
	}
	control.mu.Lock()
	control.keepWorking = false
	control.mu.Unlock()
	<-done
}
