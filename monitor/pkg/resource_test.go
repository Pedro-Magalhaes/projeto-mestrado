package coord

import (
	"reflect"
	"testing"
)

func TestCreateResource(t *testing.T) {
	type args struct {
		path    string
		project string
	}
	tests := []struct {
		name    string
		args    args
		want    *Resource
		wantErr bool
	}{
		{"one name after jobid", args{"a/b", "projectx"}, &Resource{0, "a/b", "a", "projectx"}, false},
		{"Path without slath", args{"a", "projectx"}, &Resource{0, "a", "a", "projectx"}, false},
		{"Path with multiple slashes", args{"a/b/c/d", "projectx"}, &Resource{0, "a/b/c/d", "a", "projectx"}, false},
		{"Empty path", args{"", "projectx"}, nil, true},
		{"Path with spaces only", args{"  \t\n", "projectx"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateResource(tt.args.path, tt.args.project)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResource_GetPath(t *testing.T) {
	type fields struct {
		offset int64
		path   string
		jobid  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"Retorna o campo correto", fields{0, "path", "jobid"}, "path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resource{
				offset: tt.fields.offset,
				path:   tt.fields.path,
				jobid:  tt.fields.jobid,
			}
			if got := r.GetPath(); got != tt.want {
				t.Errorf("Resource.GetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResource_GetJobid(t *testing.T) {
	type fields struct {
		offset int64
		path   string
		jobid  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"Retorna o campo correto", fields{0, "path", "jobid"}, "jobid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resource{
				offset: tt.fields.offset,
				path:   tt.fields.path,
				jobid:  tt.fields.jobid,
			}
			if got := r.GetJobid(); got != tt.want {
				t.Errorf("Resource.GetJobid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResource_GetCurrentOffset(t *testing.T) {
	type fields struct {
		offset int64
		path   string
		jobid  string
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{"Retorna o campo correto", fields{0, "path", "jobid"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resource{
				offset: tt.fields.offset,
				path:   tt.fields.path,
				jobid:  tt.fields.jobid,
			}
			if got := r.GetCurrentOffset(); got != tt.want {
				t.Errorf("Resource.GetCurrentOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}
