package consumer

import (
	"errors"
	"strings"
)

type Resource struct {
	offset  int64
	path    string
	jobid   string
	project string
}

func CreateResource(path, project, jobid string) (*Resource, error) {
	if strings.EqualFold(strings.TrimSpace(path), "") {
		return nil, errors.New("empty-path")
	}

	return &Resource{0, path, jobid, project}, nil
}

func (r *Resource) GetPath() string {
	return r.path
}

func (r *Resource) GetJobid() string {
	return r.jobid
}

func (r *Resource) GetCurrentOffset() int64 {
	return r.offset
}

func (r *Resource) SetCurrentOffset(o int64) {
	r.offset = o
}
