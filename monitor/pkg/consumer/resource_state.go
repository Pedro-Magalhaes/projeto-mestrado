/*
	Pacote que vai implementar uma forma de impedir a concorrencia na hora de manipular o estado
	dos recursos sendo observados. Baseado no padrão: https://gobyexample.com/stateful-goroutines
	utiliza canais de comunicação e uma única rotina que vai sincronizar os pedidos de acesso à
	memória

	Autor: Pedro Magalhães
*/
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
var unlock chan bool
var stop chan bool

/*
	Função que vai rodar dentro de uma rotina e vai garantir que os acessos á variavel "state" é feita sem
	problemas de concorrência. Utiliza os canais para decidir que tipo de acesso é desejado e repassa para
	os handlers quando necessário(put,get e delete). O lock e unlock são usados por funções do modulo que
	fazer o acesso ao state diretamente e conseguem com esses canais bloquear/desbloquear outros acessos.
*/
func stateHandle() {
	for {
		select {
		case <-lock:
			<-unlock // wait for read to "unlock"
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

/*
	Função que vai tratar o Delete, é chamada pela rotina que está rodando o stateHandle.
	Recebe como parametro uma mensagem do tipo "msg" e pode fazer 3 tipos de deleção dependendo
	do valor dos campos da mensagem.
	1) remover partição -> jKey é a string vazia e a partição exite em "state"
		torna state[partição] = nil e fecha o canal da partição se ele existir para que todos os
		recursos vinculados à partição sejam interrompidos.
	2) remover job -> rKey é a string vazia, jKey é uma string não vazia e existe em state[partição]
	e a partição exite em "state"
		torna state[partição][jKey] = nil e fecha o canal do job se ele existir para que todos os
		recursos vinculados ao job sejam interrompidos.
	3) *Não implementado* rKey não é vazia e existe em state[partição][jKey], jKey não é vazia a e existe em state[partição]
	e a partição exite em "state" ---- Essa função não foi implementada porque está sendo tratado diretamente
	no consumer e sendo atualizado via PutResource
*/
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
		}
		// else { // delete resource
		// 	 state[m.partition][m.jKey] = nil
		// 	 close(jobChan[m.jKey])
		// 	 jobChan[m.jKey] = nil
		// }
	}
}

/*
	Função que insere recursos no "state". Assim como a função handleDelete é usada na rotina do stateHandle
	e recebe além da partição, jKey e rKey, o resourseState de um recurso. Assume que todos os parametros
	estão corretos  jKey e rKey não vazias
	Se necessário cria maps que ainda não foram inicializados como quando inserimos o primeiro elemento em
	uma partição ou o primeiro recurso dentro da chave jKey
*/
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

/*
	Função que retorna recursos do "state". Assim como a função handleDelete é usada na rotina do stateHandle
	Rebebe assim como as outras funções anteriores partition, jKey e rKey e tem 2 comportamentos, podendo
	retornar um ResourceState ou um map da partição dependendo dos parametros.
	1) retorna o mapa da partição: Se a partição existe e o jKey é a string vazia
	2) retorna o ResourceState: Se a partição existe, a chave jKey existe na partição e o rKey existe
	3) retona nulo em outros casos
	O retorno é feito no mesmo canal que é usado no stateHandle para o get
*/
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

// Função que para a rotina do rodando a função stateHandle é usada na função StopStateStore
func sendStopSignal() {
	stop <- true
}

// função que inicializa as variáveis do modulo e da start na rotina que vai rodar o stateHandle
// é usada somente pela função StartStateStore
func initializeState() {
	log.Println("StateStore: initializeState")
	state = make(map[int32]map[string]map[string]*ResourceState)
	jobChan = make(map[string]chan bool)
	partitionChan = make(map[int32]chan bool)
	put = make(chan putMsg)
	get = make(chan getMsg)
	delete = make(chan msg)
	lock = make(chan bool)
	unlock = make(chan bool)
	stop = make(chan bool)
	go stateHandle()
}

// Função publica que faz o initializeState, usa o Once.Do para garantir que a inicialização
// só vai ocorrer uma vez
func StartStateStore() {
	log.Println("StateStore: Starting Routine")
	initOnce.Do(initializeState)
}

// Função publica que faz a finalização do módulo, usa o Once.Do para garantir que a finalização
// só vai ocorrer uma vez
func StopStateStore() {
	log.Println("StateStore: Stoping Routine")
	endOnce.Do(sendStopSignal)
}

/*
	Retorna o channel de um determinado job
	recebe a string jobid que vai servir como mapa. Assume que o parametro é diferente de ""
*/
func GetJobChan(jobid string) chan bool {
	return jobChan[jobid]
}

/*
	Retorna o channel de uma determinada partição
	recebe a string jobid que vai servir como mapa. Assume que o parametro é válido
*/
func GetPartitionChan(p int32) chan bool {
	return partitionChan[p]
}

/*
	Função pública que vai enviar mensagem para o stateHandle e obter um determinado ResourceState
	recebe a partição, a chave do job e a chave do recurso
	Retorna o que for retornado na mensagem
*/
func GetResource(partition int32, jobKey, resourceKey string) *ResourceState {
	log.Println("StateStore: GetResource", partition, jobKey, resourceKey)
	defer log.Println("StateStore: GetResource END")
	get <- getMsg{msg: msg{partition: partition, jKey: jobKey, rKey: resourceKey}}
	m := <-get
	return m.singleItem
}

/*
	Função pública que vai enviar mensagem para o stateHandle para atualizar um determinado ResourceState
	recebe a partição, a chave do job e a chave do recurso e o ResourceState que vai ser colocado na posição
*/
func PutResource(partition int32, jobKey, resourceKey string, r *ResourceState) {
	log.Println("StateStore: PutResource", partition, jobKey, resourceKey)
	defer log.Println("StateStore: PutResource END")
	put <- putMsg{r: r, msg: msg{jKey: jobKey, rKey: resourceKey, partition: partition}}
}

/*
	Função pública que vai enviar mensagem para o stateHandle para deletar uma partição
	recebe a partição a ser deletada
*/
func DeletePartition(p int32) {
	delete <- msg{partition: p}
}

/*
	Função pública que vai enviar mensagem para o stateHandle para deletar um jobid de uma partição
	recebe a partição e a chave do job a ser deletada
*/
func DeleteJob(p int32, jobid string) {
	delete <- msg{partition: p, jKey: jobid}
}

/*
	Função publica que retorna as partições existentes no mapa "state"
*/
func GetPartitions() []int32 {
	lock <- true // getting lock
	var partitions []int32
	for p, v := range state {
		if v != nil {
			partitions = append(partitions, p)
		}
	}
	unlock <- true // unlock
	return partitions
}

// Função que retorna todos os recursos de uma deteminada partição
// é usada para atualizar o estado das partições quando algum recurso é inserido ou deletado
// recebe por parametro a partição e retorna a lista de todos os recursos existentes nela
// Melhoria futura: pode ser necessário deixar em memoria a lista de recursos (se ficar lento)
func GetPartitionResources(p int32) []Resource {
	lock <- true // getting lock
	var resources []Resource
	for _, jobResouces := range state[p] {
		for _, v := range jobResouces {
			resources = append(resources, *v.R)
		}
	}
	unlock <- true // unlock
	return resources
}
