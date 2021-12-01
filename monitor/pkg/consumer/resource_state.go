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

type putMsg struct {
	r          *ResourceState
	jKey, rKey string
	partition  int32
}

type getMsg struct {
	partitionMap map[string]map[string]*ResourceState
	singleItem   *ResourceState
	rKey, jKey   string
	partition    int32
}

var state map[int32]map[string]map[string]*ResourceState
var initOnce sync.Once
var endOnce sync.Once

var put chan putMsg
var get chan getMsg
var lock chan bool
var stop chan bool

func stateHandle() {
	for {
		select {
		case <-lock:
			lock <- true // wait for read to "unlock"
		case <-stop:
			log.Println("Stopping state handler")
			return
		case m := <-get:
			handleGet(m)
		case m := <-put:
			handlePut(m)
		}
	}
}

func handlePut(m putMsg) {
	log.Printf("StateStore: handlePut. MSG: %+v\n", m)
	if state[m.partition] == nil {
		state[m.partition] = make(map[string]map[string]*ResourceState)
	}
	jMap := state[m.partition]
	if rMap := jMap[m.jKey]; rMap == nil {
		jMap[m.jKey] = make(map[string]*ResourceState)
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
	put = make(chan putMsg)
	get = make(chan getMsg)
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

func GetResource(partition int32, jobKey, resourceKey string) *ResourceState {
	log.Println("StateStore: GetResource", partition, jobKey, resourceKey)
	defer log.Println("StateStore: GetResource END")
	get <- getMsg{partition: partition, jKey: jobKey, rKey: resourceKey}
	m := <-get
	return m.singleItem
}

// // JobResourceMap
// func GetPartitionMap(partition int32) map[string]map[string]*ResourceState {
// 	log.Println("StateStore: GetPartitionMap", partition)
// 	defer log.Println("StateStore: GetPartitionMap END")
// 	get <- getMsg{partition: partition, jKey: ""}
// 	m := <-get
// 	return m.partitionMap
// }

func PutResource(partition int32, jobKey, resourceKey string, r *ResourceState) {
	log.Println("StateStore: PutResource", partition, jobKey, resourceKey)
	defer log.Println("StateStore: PutResource END")
	put <- putMsg{r: r, jKey: jobKey, rKey: resourceKey, partition: partition}
}

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
