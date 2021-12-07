package consumer

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func init() {
	StartStateStore()
}

func Test_stateHandle(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Lock order: 1"},
		{name: "Lock order: 2"},
		{name: "Lock order: 3"},
		{name: "Lock order: 4"},
		{name: "Lock order: 5"},
		{name: "Lock order: 6"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			lock <- true // lock state
			go func() {
				GetPartitions()
			}()

			go func() {
				PutResource(0, "jobid", "resourceId", &ResourceState{})
			}()

			go func() {
				DeletePartition(0)
			}()

			time.Sleep(100 * time.Millisecond) // sleep pra garantir que as rotinas estão prontas
			unlock <- true                     // unlock
			time.Sleep(100 * time.Millisecond) // sleep pra garantir que as rotinas terminaram
			// esse teste só serve se usarmos a flag "-race" para que go detecte condição de corrida
		})
	}
}

func Test_handleDelete(t *testing.T) {
	p := int32(0)
	jobid1 := "jobid1"
	jobid2 := "jobid2"
	resourceId := "rid"
	testeState := ResourceState{BeeingWatched: true}
	noJobSample := make(map[int32]map[string]map[string]*ResourceState)
	completeStateSample := make(map[int32]map[string]map[string]*ResourceState)
	completeStateSample[p] = make(map[string]map[string]*ResourceState)
	completeStateSample[p][jobid1] = make(map[string]*ResourceState)
	completeStateSample[p][jobid1][resourceId] = &testeState
	completeStateSample[p][jobid2] = make(map[string]*ResourceState)
	completeStateSample[p][jobid2][resourceId] = &testeState
	onlyJob2 := noJobSample
	onlyJob2[p] = make(map[string]map[string]*ResourceState)
	onlyJob2[p][jobid2] = completeStateSample[p][jobid2]
	completeStateSampleNoRecJob1 := completeStateSample
	completeStateSample[p][jobid1][resourceId] = nil
	// nilSample := make(map[int32]map[string]map[string]*ResourceState)
	noJobSample[p] = nil

	type args struct {
		m msg
	}
	tests := []struct {
		name   string
		args   args
		before map[int32]map[string]map[string]*ResourceState
		expect map[int32]map[string]map[string]*ResourceState
	}{
		{"Deleta partição", args{m: msg{jKey: "", rKey: "", partition: 0}}, completeStateSample, noJobSample},
		{"Deleta job1", args{m: msg{jKey: jobid1, rKey: "", partition: 0}}, completeStateSample, onlyJob2},
		{"Deleta recurso", args{m: msg{jKey: jobid1, rKey: resourceId, partition: 0}}, completeStateSample, completeStateSampleNoRecJob1},
		{"Deleta recurso que não existe", args{m: msg{jKey: jobid1, rKey: "naoExiste", partition: 0}}, completeStateSample, completeStateSample},
		{"Deleta job1 quando ele não existe", args{m: msg{jKey: jobid1, rKey: "", partition: 0}}, onlyJob2, onlyJob2},
		{"Deleta partição que não existe", args{m: msg{jKey: "", rKey: "", partition: 10}}, completeStateSample, completeStateSample},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lock <- true
			state = tt.before
			unlock <- true
			handleDelete(tt.args.m)
			if !reflect.DeepEqual(state, tt.expect) {
				t.Errorf("Error. Expected: %#v, Got: %#v\n", tt.expect, state)
			}
		})
	}
}

func Test_handlePut(t *testing.T) {
	p := int32(0)
	jobid1 := "jobid1"
	jobid2 := "jobid2"
	resourceId := "rid"
	testeState := ResourceState{BeeingWatched: true}
	completeStateSample := make(map[int32]map[string]map[string]*ResourceState)
	completeStateSample[p] = make(map[string]map[string]*ResourceState)
	completeStateSample[p][jobid1] = make(map[string]*ResourceState)
	completeStateSample[p][jobid1][resourceId] = &testeState
	completeStateSample[p][jobid2] = make(map[string]*ResourceState)
	completeStateSample[p][jobid2][resourceId] = &testeState
	onlyJob1 := make(map[int32]map[string]map[string]*ResourceState)
	onlyJob1[p] = make(map[string]map[string]*ResourceState)
	onlyJob1[p][jobid1] = make(map[string]*ResourceState)
	onlyJob1[p][jobid1][resourceId] = &testeState
	onlyJob2 := make(map[int32]map[string]map[string]*ResourceState)
	onlyJob2[p] = make(map[string]map[string]*ResourceState)
	onlyJob2[p][jobid2] = completeStateSample[p][jobid2]
	// nilSample := make(map[int32]map[string]map[string]*ResourceState)
	noJobSample := make(map[int32]map[string]map[string]*ResourceState)
	noJobSample[p] = nil
	type args struct {
		m putMsg
	}
	tests := []struct {
		name   string
		args   args
		before map[int32]map[string]map[string]*ResourceState
		expect map[int32]map[string]map[string]*ResourceState
	}{
		{"insere partição", args{m: putMsg{msg: msg{jKey: jobid1, rKey: resourceId, partition: 0}, r: &testeState}}, noJobSample, onlyJob1},
		{"insere job1", args{m: putMsg{msg: msg{jKey: jobid1, rKey: resourceId, partition: 0}, r: &testeState}}, onlyJob2, completeStateSample},
		{"insere recurso quando já existia", args{m: putMsg{msg: msg{jKey: jobid1, rKey: resourceId, partition: 0}, r: &testeState}}, completeStateSample, completeStateSample},
		// {"insere partição que não existia", args{m: putMsg{msg: msg{jKey: jobid1, rKey: resourceId, partition: 0}, r: &testeState}}, nilSample, onlyJob1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lock <- true
			state = tt.before
			unlock <- true
			handlePut(tt.args.m)
			if !reflect.DeepEqual(state, tt.expect) {
				t.Errorf("Error. Expected: %#v, Got: %#v, Before: %#v\n", tt.expect, state, tt.before)
			}
		})
	}
}

// teste será feito no método publico
func Test_handleGet(t *testing.T) {
	type args struct {
		m getMsg
	}
	tests := []struct {
		name   string
		args   args
		before map[int32]map[string]map[string]*ResourceState
		got    *ResourceState
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
	p1 := int32(0)
	p2 := int32(1)
	jobid1 := "jobid1"
	jobid2 := "jobid2"
	resourceId1 := "rid1"
	resourceId2 := "rid2"
	testeState1 := ResourceState{BeeingWatched: true}
	testeState2 := ResourceState{BeeingWatched: true, CreatingWatcher: true}
	emptyState := make(map[int32]map[string]map[string]*ResourceState)
	completeStateSample := make(map[int32]map[string]map[string]*ResourceState)
	completeStateSample[p1] = make(map[string]map[string]*ResourceState)
	completeStateSample[p1][jobid1] = make(map[string]*ResourceState)
	completeStateSample[p1][jobid1][resourceId1] = &testeState1
	completeStateSample[p1][jobid2] = make(map[string]*ResourceState)
	completeStateSample[p1][jobid2][resourceId2] = &testeState2
	completeStateSample[p2] = make(map[string]map[string]*ResourceState)
	completeStateSample[p2][jobid1] = make(map[string]*ResourceState)
	completeStateSample[p2][jobid1][resourceId1] = &testeState1
	completeStateSample[p2][jobid2] = make(map[string]*ResourceState)
	completeStateSample[p2][jobid2][resourceId2] = &testeState2
	type args struct {
		partition   int32
		jobKey      string
		resourceKey string
	}
	tests := []struct {
		name   string
		args   args
		before map[int32]map[string]map[string]*ResourceState
		want   *ResourceState
	}{
		{"get p1 job1 resource1", args{jobKey: jobid1, resourceKey: resourceId1, partition: p1}, completeStateSample, &testeState1},
		{"get p2 job1 resource2", args{jobKey: jobid2, resourceKey: resourceId2, partition: p2}, completeStateSample, &testeState2},
		{"get from inxistent partition", args{jobKey: jobid1, resourceKey: resourceId2, partition: 10}, completeStateSample, nil},
		{"get from inxistent job", args{jobKey: "notFound", resourceKey: resourceId1, partition: p1}, completeStateSample, nil},
		{"get from inxistent resouce", args{jobKey: jobid1, resourceKey: "notFound", partition: p1}, completeStateSample, nil},
		{"get from empty state", args{jobKey: jobid1, resourceKey: resourceId1, partition: p1}, emptyState, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldState := state
			lock <- true
			state = tt.before
			unlock <- true
			if got := GetResource(tt.args.partition, tt.args.jobKey, tt.args.resourceKey); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResource() = %v, want %v", got, tt.want)
			}
			state = oldState
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
	p1 := int32(0)
	p2 := int32(1)
	p3 := int32(2)
	emptyState := make(map[int32]map[string]map[string]*ResourceState)
	completeStateSample := make(map[int32]map[string]map[string]*ResourceState)
	completeStateSample[p1] = make(map[string]map[string]*ResourceState)
	completeStateSample[p2] = make(map[string]map[string]*ResourceState)
	completeStateSample[p3] = make(map[string]map[string]*ResourceState)
	twoPartitionsState := make(map[int32]map[string]map[string]*ResourceState)
	twoPartitionsState[p1] = make(map[string]map[string]*ResourceState)
	twoPartitionsState[p3] = make(map[string]map[string]*ResourceState)
	tests := []struct {
		name   string
		before map[int32]map[string]map[string]*ResourceState
		want   []int32
	}{
		{name: "Three partitions", before: completeStateSample, want: []int32{p1, p2, p3}},
		{name: "Two partitions", before: twoPartitionsState, want: []int32{p1, p3}},
		{name: "empty state", before: emptyState, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldState := state
			lock <- true
			state = tt.before
			unlock <- true
			got := GetPartitions()
			sort.SliceStable(got, func(i, j int) bool {
				return got[i] < got[j]
			})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPartitions() = %v, want %v", got, tt.want)
			}
			state = oldState
		})
	}
}

func TestGetPartitionResources(t *testing.T) {
	p1 := int32(0)
	p2 := int32(1)
	jobid1 := "jobid1"
	jobid2 := "jobid2"
	resourceId1 := "rid1"
	resourceId2 := "rid2"
	r1 := Resource{Path: "r1"}
	r2 := Resource{Path: "r2"}
	r3 := Resource{Path: "r3"}
	testeState1 := ResourceState{BeeingWatched: true, R: &r1}
	testeState2 := ResourceState{BeeingWatched: true, CreatingWatcher: true, R: &r2}
	testeState3 := ResourceState{BeeingWatched: true, CreatingWatcher: true, R: &r3}

	emptyState := make(map[int32]map[string]map[string]*ResourceState)
	completeStateSample := make(map[int32]map[string]map[string]*ResourceState)
	completeStateSample[p1] = make(map[string]map[string]*ResourceState)
	completeStateSample[p1][jobid1] = make(map[string]*ResourceState)
	completeStateSample[p1][jobid1][resourceId1] = &testeState1
	completeStateSample[p1][jobid2] = make(map[string]*ResourceState)
	completeStateSample[p1][jobid2][resourceId2] = &testeState2
	completeStateSample[p2] = make(map[string]map[string]*ResourceState)
	completeStateSample[p2][jobid1] = make(map[string]*ResourceState)
	completeStateSample[p2][jobid1][resourceId1] = &testeState3
	type args struct {
		p int32
	}
	tests := []struct {
		name   string
		args   args
		before map[int32]map[string]map[string]*ResourceState
		want   []Resource
	}{
		{name: "p1 from complete", args: args{p: p1}, before: completeStateSample, want: []Resource{r1, r2}},
		{name: "p2 from complete", args: args{p: p2}, before: completeStateSample, want: []Resource{r3}},
		{name: "empty state", args: args{p: p1}, before: emptyState, want: nil},
	}
	for _, tt := range tests {
		oldState := state
		lock <- true
		state = tt.before
		unlock <- true
		t.Run(tt.name, func(t *testing.T) {
			got := GetPartitionResources(tt.args.p)
			sort.SliceStable(got, func(i, j int) bool {
				return got[i].Path < got[j].Path
			})
			sort.SliceStable(tt.want, func(i, j int) bool {
				return tt.want[i].Path < tt.want[j].Path
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPartitionResources() = %v, want %v", got, tt.want)
			}
		})
		state = oldState
	}
}
