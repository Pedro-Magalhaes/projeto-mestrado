package consumer

import (
	"reflect"
	"testing"
)

func Test_stateHandle(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateHandle()
		})
	}
}

func Test_handleDelete(t *testing.T) {
	type args struct {
		m msg
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleDelete(tt.args.m)
		})
	}
}

func Test_handlePut(t *testing.T) {
	type args struct {
		m putMsg
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlePut(tt.args.m)
		})
	}
}

func Test_handleGet(t *testing.T) {
	type args struct {
		m getMsg
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleGet(tt.args.m)
		})
	}
}

func Test_sendStopSignal(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sendStopSignal()
		})
	}
}

func Test_initializeState(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initializeState()
		})
	}
}

func TestStartStateStore(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			StartStateStore()
		})
	}
}

func TestStopStateStore(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			StopStateStore()
		})
	}
}

func TestGetJobChan(t *testing.T) {
	type args struct {
		jobid string
	}
	tests := []struct {
		name string
		args args
		want chan bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetJobChan(tt.args.jobid); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetJobChan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPartitionChan(t *testing.T) {
	type args struct {
		p int32
	}
	tests := []struct {
		name string
		args args
		want chan bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPartitionChan(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPartitionChan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetResource(t *testing.T) {
	type args struct {
		partition   int32
		jobKey      string
		resourceKey string
	}
	tests := []struct {
		name string
		args args
		want *ResourceState
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetResource(tt.args.partition, tt.args.jobKey, tt.args.resourceKey); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPutResource(t *testing.T) {
	type args struct {
		partition   int32
		jobKey      string
		resourceKey string
		r           *ResourceState
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PutResource(tt.args.partition, tt.args.jobKey, tt.args.resourceKey, tt.args.r)
		})
	}
}

func TestDeletePartition(t *testing.T) {
	type args struct {
		p int32
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeletePartition(tt.args.p)
		})
	}
}

func TestDeleteJob(t *testing.T) {
	type args struct {
		p     int32
		jobid string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteJob(tt.args.p, tt.args.jobid)
		})
	}
}

func TestGetPartitions(t *testing.T) {
	tests := []struct {
		name string
		want []int32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPartitions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPartitions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPartitionResources(t *testing.T) {
	type args struct {
		p int32
	}
	tests := []struct {
		name string
		args args
		want []Resource
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPartitionResources(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPartitionResources() = %v, want %v", got, tt.want)
			}
		})
	}
}
