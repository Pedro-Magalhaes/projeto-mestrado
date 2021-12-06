package consumer

import (
	"log"
	"sync"
)

type ResourceState struct {
	CreatingWatcher bool
	BeeingWatched   bool
	KeepWorking     *chan bool
	R               *Resource
}

type ResourceMap map[string]*ResourceState

type JobResourceMap map[string]map[string]ResourceMap

type msg struct {
	jKey, rKey string
	partition  int32
}
type putMsg struct {
	r *ResourceState
	msg
}

type getMsg struct {
	partitionMap map[string]map[string]*ResourceState
	singleItem   *ResourceState
	msg
}

var state map[int32]map[string]map[string]*ResourceState
var jobChan map[string]chan bool
var partitionChan map[int32]chan bool
var initOnce sync.Once
var endOnce sync.Once

var put chan putMsg
var get chan getMsg
var delete chan msg
var lock chan bool
var stop chan bool

func stateHandle() {
	for {
		select {
		case <-lock:
			lock <- true // wait for read to "unlock"
		case <-stop:
			log.Println("Resource State: Stopping state handler")
			return
		case m := <-get:
			handleGet(m)
		case m := <-put:
			handlePut(m)
		case m := <-delete:
			handleDelete(m)
		}
	}
}

func handleDelete(m msg) {
	if state[m.partition] != nil {
		if m.jKey == "" { // delete pattition
			log.Printf("State Store: deleting pattition %d.\n", m.partition)
			state[m.partition] = nil
			if partitionChan[m.partition] != nil {
				log.Printf("State Store: closing channel from partition %d.\n", m.partition)
				close(partitionChan[m.partition]) //close chan, all resources from partitions will stop
				partitionChan[m.partition] = nil
			}
		} else if m.rKey == "" { // delete job
			if state[m.partition][m.jKey] != nil {
				log.Printf("State Store: deleting job %s.\n", m.jKey)
				state[m.partition][m.jKey] = nil
				if jobChan[m.jKey] != nil {
					log.Printf("State Store: closing channel from job %s.\n", m.jKey)
					close(jobChan[m.jKey])
					jobChan[m.jKey] = nil
				}
			}
		} else { // delete resource
			// state[m.partition][m.jKey] = nil
			// close(jobChan[m.jKey])
			// jobChan[m.jKey] = nil
		}
	}
}

func handlePut(m putMsg) {
	log.Printf("StateStore: handlePut. MSG: %+v\n", m)
	if state[m.partition] == nil {
		state[m.partition] = make(map[string]map[string]*ResourceState)
		partitionChan = make(map[int32]chan bool)
		partitionChan[m.partition] = make(chan bool)
	}
	jMap := state[m.partition]
	if rMap := jMap[m.jKey]; rMap == nil {
		jMap[m.jKey] = make(map[string]*ResourceState)
		jobChan[m.jKey] = make(chan bool)
	}
	state[m.partition][m.jKey][m.rKey] = m.r
	log.Printf("StateStore: Storing resource. MSG: %+v\n", m.r)
}

func handleGet(m getMsg) {
	log.Printf("StateStore: handleGet. MSG: %+v\n", m)
	if state[m.partition] == nil {
		m.partitionMap = nil
	} else if m.jKey == "" {
		m.partitionMap = state[m.partition]
	} else {
		if state[m.partition][m.jKey] != nil && state[m.partition][m.jKey][m.rKey] != nil {
			m.singleItem = state[m.partition][m.jKey][m.rKey]
		}
	}
	get <- m
}

func sendStopSignal() {
	stop <- true
}

func initializeState() {
	log.Println("StateStore: initializeState")
	state = make(map[int32]map[string]map[string]*ResourceState)
	jobChan = make(map[string]chan bool)
	partitionChan = make(map[int32]chan bool)
	put = make(chan putMsg)
	get = make(chan getMsg)
	delete = make(chan msg)
	lock = make(chan bool)
	stop = make(chan bool)
	go stateHandle()
}

func StartStateStore() {
	log.Println("StateStore: Starting Routine")
	initOnce.Do(initializeState)
}

func StopStateStore() {
	log.Println("StateStore: Stoping Routine")
	endOnce.Do(sendStopSignal)
}

func GetJobChan(jobid string) chan bool {
	return jobChan[jobid]
}

func GetPartitionChan(p int32) chan bool {
	return partitionChan[p]
}

func GetResource(partition int32, jobKey, resourceKey string) *ResourceState {
	log.Println("StateStore: GetResource", partition, jobKey, resourceKey)
	defer log.Println("StateStore: GetResource END")
	get <- getMsg{msg: msg{partition: partition, jKey: jobKey, rKey: resourceKey}}
	m := <-get
	return m.singleItem
}

func PutResource(partition int32, jobKey, resourceKey string, r *ResourceState) {
	log.Println("StateStore: PutResource", partition, jobKey, resourceKey)
	defer log.Println("StateStore: PutResource END")
	put <- putMsg{r: r, msg: msg{jKey: jobKey, rKey: resourceKey, partition: partition}}
}

func DeletePartition(p int32) {
	delete <- msg{partition: p}
}

func DeleteJob(p int32, jobid string) {
	delete <- msg{partition: p, jKey: jobid}
}

// func DeletePartition()  {

// }

func GetPartitions() []int32 {
	lock <- true // getting lock
	var partitions []int32
	for p, v := range state {
		if v != nil {
			partitions = append(partitions, p)
		}
	}
	<-lock // unlock
	return partitions
}

// vai ser usado quando perdemos uma partição
// dado uma partição eu pego a lista de recursos
// investigar se posso deixar em memoria a lista de recursos (se ficar lento)
func GetPartitionResources(p int32) []Resource {
	lock <- true // getting lock
	var resources []Resource
	for _, jobResouces := range state[p] {
		for _, v := range jobResouces {
			resources = append(resources, *v.R)
		}
	}
	<-lock // unlock
	return resources
}
