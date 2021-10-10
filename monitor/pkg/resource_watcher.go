package coord

import (
	"io"
	"log"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type safeBool struct {
	mu   sync.Mutex
	work bool
}

// receive a array of bytes return a boolean
// the boolean will control the offset "commit"
type watchCallback func([]byte) bool

func WatchResource(r *Resource, control *safeBool, maxChunkSize uint, cb watchCallback, state *ResourceSafeMap) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Printf("Error creating new Watcher:  " + err.Error())
		setResourceState(state, r, resourceState{false, false})
		return
	}
	setResourceState(state, r, resourceState{beeingWatched: true})
	log.Println("Setando true para watcher")
	defer watcher.Close()
	keepWorking := true
	done := make(chan bool)
	go func() {
		log.Printf("Iniciando loop do watcher")
		for {
			control.mu.Lock()
			keepWorking = control.work
			control.mu.Unlock()
			if !keepWorking {
				log.Println("Watcher: Got stopped by control variable")
				break
			}

			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("ERROR: watcher.Events ", event)
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					err := handleFileChange(r, control, maxChunkSize, cb, state)
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
		setResourceState(state, r, resourceState{false, false})
		done <- true
	}()

	err = watcher.Add(r.path)
	if err != nil {
		log.Println("error:", err)
		control.mu.Lock()
		control.work = false
		control.mu.Unlock()
		setResourceState(state, r, resourceState{false, false})
		return
	}
	<-done
}

func handleFileChange(r *Resource, control *safeBool, maxChunkSize uint, cb watchCallback, state *ResourceSafeMap) *error {
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
		keepSending := cb(buffer[:bytesRead])
		if !keepSending {
			break
		}
		currOffset += int64(bytesRead)
	}
	r.offset = currOffset
	return nil
}

func setResourceState(state *ResourceSafeMap, r *Resource, resourceState resourceState) {
	state.mu.Lock()
	state.resourceMap[r.path] = resourceState
	state.mu.Unlock()
}
