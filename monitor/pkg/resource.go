package coord

import (
	"strings"
)

type Resource struct {
	path  string
	jobid string
}

func CreateResource(path string) *Resource {
	if strings.EqualFold(path, "") {
		return nil
	}
	paths := strings.Split(path, "/")

	return &Resource{
		path:  path,
		jobid: paths[0],
	}
}

func (r Resource) GetPath() string {
	return r.path
}

func (r Resource) GetJobid() string {
	return r.jobid
}
