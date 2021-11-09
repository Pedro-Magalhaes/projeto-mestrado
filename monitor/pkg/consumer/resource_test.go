package consumer

import (
	"reflect"
	"testing"
)

func TestCreateResource(t *testing.T) {
	type args struct {
		path    string
		project string
		jobid   string
	}
	tests := []struct {
		name    string
		args    args
		want    *Resource
		wantErr bool
	}{
		{"one name after jobid", args{"a/b", "projectx", "jobid1"}, &Resource{0, "a/b", "jobid1", "projectx"}, false},
		{"Path without slath", args{"a", "projectx", "jobid1"}, &Resource{0, "a", "jobid1", "projectx"}, false},
		{"Path with multiple slashes", args{"a/b/c/d", "projectx", "jobid1"}, &Resource{0, "a/b/c/d", "jobid1", "projectx"}, false},
		{"Empty path", args{"", "projectx", "jobid1"}, nil, true},
		{"Path with spaces only", args{"  \t\n", "projectx", "jobid1"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateResource(tt.args.path, tt.args.project, tt.args.jobid)
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

func TestResource_SetCurrentOffset(t *testing.T) {
	type fields struct {
		offset  int64
		path    string
		jobid   string
		project string
	}
	type args struct {
		o int64
	}
	type want struct {
		o int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{"Retorna o campo correto", fields{0, "path", "jobid", "p"}, args{1}, want{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resource{
				offset:  tt.fields.offset,
				path:    tt.fields.path,
				jobid:   tt.fields.jobid,
				project: tt.fields.project,
			}
			r.SetCurrentOffset(tt.args.o)
			if got := r.GetCurrentOffset(); got != tt.want.o {
				t.Errorf("Resource.GetCurrentOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}
