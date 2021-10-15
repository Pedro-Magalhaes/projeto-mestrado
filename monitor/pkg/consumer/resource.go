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

func CreateResource(path string, project string) (*Resource, error) {
	if strings.EqualFold(strings.TrimSpace(path), "") {
		return nil, errors.New("empty-path")
	}
	paths := strings.Split(path, "/")

	return &Resource{0, path, paths[0], project}, nil
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
