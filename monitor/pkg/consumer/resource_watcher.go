/*
	Pacote que implementa a monitoração de arquivos. Vai criar a monitoração e sempre que houver escrita
	vai enviar os novos bytes escritos para uma callback recebida. Assume que o arquivo sempre sofre "append"
	ou seja ele só cresce e nada é deletado. Se

	Autor: Pedro Magalhães
*/
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

/*
	Função que cria inicia a rotina de observação de um arquivo baseado em um "Resource"
	sempre que é chamada essa função cria um nova rotina diferente, ou seja, cada arquivo é
	observado por uma rotina diferente.
	Essa função é bloqueante e só retorna quando a rotina que observa o arquivo para.
*/
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
				log.Println("Stopping watcher. Job stoped", r.GetPath())
				done <- true
				return
			case <-partitionChan:
				log.Println("Stopping watcher. Partition stoped", r.GetPath())
				done <- true
				return
			case <-keepWorking:
				log.Println("Stopping watcher. ", r.GetPath())
				rs.CreatingWatcher = false
				rs.BeeingWatched = false
				setResourceState(r, rs)
				done <- true
				return
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("ERROR: watcher.Events ", event)
					done <- true
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

/*
	Função que vai ser chamada quando um arquivo sofrer uma escrita. Vai abrir o arquivo ele ler
	até o fim chamando a callback de "chunk" em "chunk". Ao final atualiza o offset
*/
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

/*
	Função auxiliar para fazer a atualização do estado do recurso
*/
func setResourceState(r *Resource, resourceState ResourceState) {
	r.PutStateToStateStore(&resourceState)
}

/*
	Função auxiliar para pegar o canal que indica se um recurso deve parar de ser observado
*/
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
