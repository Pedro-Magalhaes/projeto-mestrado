package consumer

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pfsmagalhaes/monitor/pkg/util"
)

type safeBool struct {
	mu   sync.Mutex
	work bool
}

// receive a array of bytes return a boolean
// the boolean will control the offset "commit"
type watchCallback func([]byte, int64) bool

func WatchResource(r *Resource, maxChunkSize uint, cb watchCallback, state *util.ResourceSafeMap) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Printf("Error creating new Watcher:  " + err.Error())
		setResourceState(state, r, util.ResourceState{CreatingWatcher: false, BeeingWatched: false})
		return
	}
	setResourceState(state, r, util.ResourceState{BeeingWatched: true})
	log.Println("Setando true para watcher")
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		log.Printf("Iniciando loop do watcher")
		for {
			keepWorking := getKeepWorkingChan(state, r)
			select {
			case <-keepWorking:
				log.Println("Stoping watcher. ", r.GetPath())
				setResourceState(state, r, util.ResourceState{CreatingWatcher: false, BeeingWatched: false})
				done <- true
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("ERROR: watcher.Events ", event)
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					err := handleFileChange(r, maxChunkSize, cb, state)
					if err != nil {
						break
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					break
				}
				log.Println("ERROR on Watcher:", err)
			}
		}
	}()

	err = watcher.Add(r.GetPath())
	if err != nil {
		log.Println("error:", err)
		setResourceState(state, r, util.ResourceState{CreatingWatcher: false, BeeingWatched: false})
		return
	}
	<-done
}

func handleFileChange(r *Resource, maxChunkSize uint, cb watchCallback, state *util.ResourceSafeMap) *error {
	buffer := make([]byte, maxChunkSize)
	file, err := os.Open(r.path)
	if err != nil {
		log.Println("ERROR: Could not open file: " + r.path)
		return new(error)
	}
	currOffset := r.offset
	for {
		bytesRead, err := file.ReadAt(buffer, currOffset)
		if bytesRead <= 0 || (err != nil && err != io.EOF) {
			log.Println("Stoping this chunk read, cause: " + err.Error())
			break
		}
		keepSending := cb(buffer[:bytesRead], currOffset)
		if !keepSending {
			break
		}
		currOffset += int64(bytesRead)
	}
	r.offset = currOffset
	return nil
}

func setResourceState(state *util.ResourceSafeMap, r *Resource, resourceState util.ResourceState) {
	state.Mu.Lock()
	c := state.ResourceMap[r.GetPath()].KeepWorking
	if resourceState.KeepWorking == nil {
		resourceState.KeepWorking = c
	}
	state.ResourceMap[r.GetPath()] = &resourceState
	state.Mu.Unlock()
}

// Todo: check if is safe to read state (Should I use the mutex? or use a channel instead of a boolean)
func getKeepWorkingChan(state *util.ResourceSafeMap, r *Resource) chan bool {
	state.Mu.Lock()
	c := state.ResourceMap[r.GetPath()].KeepWorking
	fmt.Println(c, r.GetPath(), state.ResourceMap[r.GetPath()])
	println(state)

	state.Mu.Unlock()
	return c
}
