/*
	Pacote que implementa um objeto do tipo Resource e encapsula sua manipulação

	Autor: Pedro Magalhães
*/
package consumer

import (
	"errors"
	"strings"
)

type Resource struct {
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
	Path      string `json:"path"`
	Jobid     string `json:"jobid"`
	Project   string `json:"project"`
}

/*
	Função que cria um nove "Resource"
	Recebe as strings path, project e jobid e o int p representando a partição do recurso
	retorna erro se o path for vazio
*/
func CreateResource(path, project, jobid string, p int32) (*Resource, error) {
	if strings.EqualFold(strings.TrimSpace(path), "") {
		return nil, errors.New("empty-path")
	}

	return &Resource{p, 0, path, jobid, project}, nil
}

/*
	Função que obtem da stateStore o estado de um recurso
*/
func (r *Resource) GetStateFromStateStore() *ResourceState {
	return GetResource(r.Partition, r.Jobid, r.GetPath())
}

/*
	Função que faz a atualização do estado de um recurso na PutStateToStateStore
	recebe com parametro o estado novo
*/
func (r *Resource) PutStateToStateStore(state *ResourceState) {
	PutResource(r.Partition, r.Jobid, r.GetPath(), state)
}

/*
	Getter para o Path
*/
func (r *Resource) GetPath() string {
	return r.Path
}

/*
	Getter para o Jobid
*/
func (r *Resource) GetJobid() string {
	return r.Jobid
}

/*
	Getter para o Offset
*/
func (r *Resource) GetCurrentOffset() int64 {
	return r.Offset
}

/*
	Setter para o Offset
*/
func (r *Resource) SetCurrentOffset(o int64) {
	r.Offset = o
}

/*
	Função que retorna o tópico correpondete ao recurso
*/
func (r *Resource) GetResourceTopic() string {
	pathArray := strings.SplitN(r.Path, "/", 1)
	index := 1
	if len(pathArray) < 2 { // no basepath
		index = 0
	}
	topic := strings.ReplaceAll(pathArray[index], "/", "__")
	return topic
}
