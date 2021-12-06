package consumer

import (
	"io"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

// receive a array of bytes return a boolean
// the boolean will control the offset "commit"
type watchCallback func([]byte, int64) bool

func WatchResource(r *Resource, maxChunkSize uint, cb watchCallback, rs ResourceState) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Printf("Error creating new Watcher:  " + err.Error())
		rs.BeeingWatched = false
		rs.CreatingWatcher = false
		setResourceState(r, rs)
		return
	}
	rs.BeeingWatched = true
	setResourceState(r, rs)
	log.Println("Setando true para watcher")
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		log.Printf("Iniciando loop do watcher")
		keepWorking := getKeepWorkingChan(r)
		jobChan := GetJobChan(rs.R.Jobid)
		partitionChan := GetPartitionChan(rs.R.Partition)
		for {
			select {
			case <-jobChan:
				log.Println("Stoping watcher. Partition stoped", r.GetPath())
				done <- true
				return
			case <-partitionChan:
				log.Println("Stoping watcher. Job stoped", r.GetPath())
				done <- true
				return
			case <-keepWorking:
				log.Println("Stoping watcher. ", r.GetPath())
				rs.CreatingWatcher = false
				rs.BeeingWatched = false
				setResourceState(r, rs)
				done <- true
				return
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("ERROR: watcher.Events ", event)
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					err := handleFileChange(r, maxChunkSize, cb)
					if err != nil {
						break
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Println("ERROR on Watcher:", err)
					break
				}
				log.Println("ERROR on Watcher:", err)
			}
		}
	}()

	err = watcher.Add(r.GetPath())
	if err != nil {
		log.Println("error:", err)
		rs.BeeingWatched = false
		rs.CreatingWatcher = false
		setResourceState(r, rs)
		return
	}
	<-done
}

func handleFileChange(r *Resource, maxChunkSize uint, cb watchCallback) *error {
	buffer := make([]byte, maxChunkSize)
	file, err := os.Open(r.GetPath())
	if err != nil {
		log.Println("ERROR: Could not open file: " + r.GetPath())
		return new(error)
	}
	currOffset := r.Offset
	for {
		bytesRead, err := file.ReadAt(buffer, currOffset)
		if bytesRead <= 0 || (err != nil && err != io.EOF) {
			log.Println("Stoping this chunk read, cause: " + err.Error())
			break
		}
		keepSending := cb(buffer[:bytesRead], currOffset)
		if !keepSending {
			log.Println("Stoping sending, callback asked to stop")
			break
		}
		currOffset += int64(bytesRead)
	}
	r.Offset = currOffset
	return nil
}

func setResourceState(r *Resource, resourceState ResourceState) {
	r.PutStateToStateStore(&resourceState)
}

func getKeepWorkingChan(r *Resource) chan bool {
	rState := r.GetStateFromStateStore()
	if rState == nil {
		log.Panicf("PANIC, state nil! %+v\nState From Store: %+v\n", r, rState)
	}
	c := r.GetStateFromStateStore().KeepWorking
	if c == nil {
		log.Panicf("PANIC, Channel nil! %+v\nState From Store: %+v\n", r, rState)
	}
	return *c
}
