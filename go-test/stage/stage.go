package stage

import (
	"log"
	"sync"
)

type Stage struct {
	Id        uint
	WaitGroup *sync.WaitGroup
	Jobs      []Job
	channel   chan bool
}

type Stages struct {
	stages map[uint]*Stage
}

type Job struct {
	Begin     *Stage
	End       *Stage
	Work      func(chan bool)
	jobChan   chan bool
	WaitGroup *sync.WaitGroup
}

func CreateStage(id uint) *Stage {
	return &Stage{
		Id:        id,
		WaitGroup: &sync.WaitGroup{},
		channel:   make(chan bool, 0),
		Jobs:      []Job{},
	}
}

func (st *Stage) runWork(j Job) {
	j.Work(j.End.channel)
	//println("Job finalizado. Check se job.WG == stage.WG", j.WaitGroup == st.WaitGroup)
	if j.Begin == j.End {
		j.WaitGroup.Done() // como colocar um WG
	}
}

// Adicona uma tarefa ao estágio
func (st *Stage) AddJob(work func(chan bool)) {
	st.AddJobMultiStage(work, st)
}

// Adiciona uma tarefa que inicia em um estágio mas termina em outro
func (st *Stage) AddJobMultiStage(work func(chan bool), endStage *Stage) *Job {
	job := Job{
		jobChan:   make(chan bool, 0),
		Work:      work,
		WaitGroup: endStage.WaitGroup,
		End:       endStage,
		Begin:     st,
	}
	if st == endStage {
		endStage.WaitGroup.Add(1)
	}
	st.Jobs = append(st.Jobs, job)
	println("Len de jobs", len(st.Jobs))
	return &job
}

// Roda um estágio
func (st *Stage) Run() {
	for _, job := range st.Jobs {
		go st.runWork(job)
	}
	st.WaitGroup.Wait()
	log.Default().Println("Finalizando stage: ", st.Id)
	// FIXME: melhorar logica, usar apenas um canal ou preciso de uma para cada job?
	for _, job := range st.Jobs {
		close(job.jobChan)
	}
	close(st.channel)
}

// Adiciona um estágio
func (s Stages) AddStage(st *Stage) {
	s.stages[st.Id] = st
}

// Adiciona um array de estágios
func (s Stages) AddStages(stages []*Stage) {
	for _, st := range stages {
		s.stages[st.Id] = st
	}
}

// retorna um estágio dado um id
func (s Stages) GetStage(id uint) *Stage {
	return s.stages[id]
}

// Roda todos os estágios internos
func (s Stages) Run() {
	for _, s := range s.stages {
		log.Default().Println("Runnig stage: ", s.Id)
		s.Run()
	}
}

// colocar um padrão usando select para jobs que não terminam e precisam ser interrompidos ao final de um "stage"
// Nesse caso não adicionamos ele em nenhum WG
