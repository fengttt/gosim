package gosim

import (
	"fmt"
	_ "log"
	"math"
)

// Logic Process, which can run at certain tic.
// Run should return
//	- positive int64, means this lp should be run again, at tic + ret
//	- 0, means put this lp to sleep.
//	- -1, means this lp should die,
type Lp interface {
	Run(s *Sim, r *Runner, msgs []interface{}) int64
}

// LP id
type Lpid struct {
	RunnerId   int
	Id         int
	CreateTick int64
}

// Sim: Simulation
type Sim struct {
	runners []Runner
	aggs    []Agg
	tick    int64
	msgs    [][]simMsg
	running bool
	ch      chan int64
}

// New creates a Sim with nr runners.
func New(nr int) *Sim {
	var sim Sim
	sim.runners = make([]Runner, nr)
	for i, _ := range sim.runners {
		sim.runners[i].init(i, nr)
	}
	sim.ch = make(chan int64)
	return &sim
}

// NumRunner returns Number of runners
func (s *Sim) NumRunner() int {
	return len(s.runners)
}

// CurrentTick returns current tick
func (s *Sim) CurrentTick() int64 {
	return s.tick
}

// AddLP adds a runner, returns lp id.
func (s *Sim) AddLp(r int, lp Lp) (Lpid, error) {
	if s.running {
		return Lpid{}, fmt.Errorf("Sim is running.")
	}

	if r >= len(s.runners) {
		return Lpid{}, fmt.Errorf("Sim does not have enough runners.")
	}
	return s.runners[r].AddLp(s, lp), nil
}

// GetLP looks up lp from lpid.
func (s *Sim) GetLp(lpid Lpid) (Lp, error) {
	if lpid.RunnerId < 0 || lpid.RunnerId >= s.NumRunner() {
		return nil, fmt.Errorf("invalid lpid")
	}
	return s.runners[lpid.RunnerId].getLp(lpid)
}

// CreateSum creates a sum agg.
func (s *Sim) CreateSum() int {
	id := len(s.aggs)
	s.aggs = append(s.aggs, newSum())
	for i, _ := range s.runners {
		s.runners[i].aggs = append(s.runners[i].aggs, newSum())
	}
	return id
}

// CreateMin creates a min agg.
func (s *Sim) CreateMin() int {
	id := len(s.aggs)
	s.aggs = append(s.aggs, newMin())
	for i, _ := range s.runners {
		s.runners[i].aggs = append(s.runners[i].aggs, newMin())
	}
	return id
}

// CreateMax creates a max agg.
func (s *Sim) CreateMax() int {
	id := len(s.aggs)
	s.aggs = append(s.aggs, newMax())
	for i, _ := range s.runners {
		s.runners[i].aggs = append(s.runners[i].aggs, newMax())
	}
	return id
}

// ReadAgg reads the value of an agg.
func (s *Sim) ReadAgg(aggId int) (int64, float64) {
	return s.aggs[aggId].Get()
}

// Run Sim, with a stop condition.
func (s *Sim) Run(stop func(s *Sim) bool) {
	for s.CurrentTick() != math.MaxInt64 && !stop(s) {
		s.runOneStep()
	}
}

// Run Sim
func (s *Sim) RunSteps(nStep int64) {
	endTick := s.CurrentTick() + nStep
	for s.CurrentTick() < endTick {
		s.runOneStep()
	}
}

func (s *Sim) runOneStep() {
	tick := s.CurrentTick()
	var nextTick int64 = math.MaxInt64

	s.running = true

	for i := 0; i < s.NumRunner(); i++ {
		s.runners[i].ch <- tick
	}

	for i := 0; i < s.NumRunner(); i++ {
		nxt := <-s.ch
		if nxt < nextTick {
			nextTick = nxt
		}
	}
	s.tick = nextTick

	// Gather aggs
	for i, _ := range s.aggs {
		for j, _ := range s.runners {
			s.aggs[i].Gather(s.runners[j].aggs[i])
			s.runners[j].aggs[i].Reset()
		}
	}

	s.running = false
}

func (s *Sim) Start() {
	for i, _ := range s.runners {
		go s.runners[i].runLoop(s)
	}
}

func (s *Sim) Stop() {
	for i := 0; i < s.NumRunner(); i++ {
		s.runners[i].ch <- -1
	}
	for i := 0; i < s.NumRunner(); i++ {
		<-s.ch
	}
}
