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

func CreateResource(path, project, jobid string, p int32) (*Resource, error) {
	if strings.EqualFold(strings.TrimSpace(path), "") {
		return nil, errors.New("empty-path")
	}

	return &Resource{p, 0, path, jobid, project}, nil
}

func (r *Resource) GetStateFromStateStore() *ResourceState {
	return GetResource(r.Partition, r.Jobid, r.GetPath())
}

func (r *Resource) PutStateToStateStore(state *ResourceState) {
	PutResource(r.Partition, r.Jobid, r.GetPath(), state)
}

func (r *Resource) GetPath() string {
	return r.Path
}

func (r *Resource) GetJobid() string {
	return r.Jobid
}

func (r *Resource) GetCurrentOffset() int64 {
	return r.Offset
}

func (r *Resource) SetCurrentOffset(o int64) {
	r.Offset = o
}

func (r *Resource) GetResourceTopic() string {
	pathArray := strings.SplitN(r.Path, "/", 1)
	index := 1
	if len(pathArray) < 2 { // no basepath
		index = 0
	}
	topic := strings.ReplaceAll(pathArray[index], "/", "__")
	return topic
}
