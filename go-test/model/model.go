package model

import "sync"

// <- chanStage3
type Stage struct {
	Id        int
	WaitGroup sync.WaitGroup
	stageChan *chan bool
}

type Job struct {
	Begin    int
	End      int // termina no 3
	Callback func()
	jobChan  *chan bool // canal do estagio 3
	endStage *Stage
}

var stageJobs map[int][]Job
var lastStage int

func initialize() {

}

func AddJob() {

}

func (s *Stage) StartStage() {
	for _, v := range stageJobs[s.Id] {
		go v.Callback()
		s.WaitGroup.Done() // o que fazer se o job terminar em outro?
	}
}
