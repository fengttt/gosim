package gosim

import (
	"fmt"
	_ "log"
	"math"
	"math/rand"
)

type lpEntry struct {
	createTick int64
	runTick    int64
	lp         Lp
	msgs       []interface{}
}

func (lpe *lpEntry) reset() {
	lpe.lp = nil
	lpe.msgs = nil
}

type simMsg struct {
	dst Lpid
	msg interface{}
}

// Runner runs LPs.
type Runner struct {
	id       int
	nextTick int64
	lpCnt    int
	lps      []lpEntry
	aggs     []Agg
	msgs     [][]simMsg
	ch       chan int64
	r        *rand.Rand
}

func (r *Runner) init(id, nr int) {
	r.id = id
	r.msgs = make([][]simMsg, nr)
	r.ch = make(chan int64)
	r.r = rand.New(rand.NewSource(int64(id)))
}

func (r *Runner) numLps() int {
	return r.lpCnt
}

func (r *Runner) Rand() *rand.Rand {
	return r.r
}

func (r *Runner) AddLp(s *Sim, lp Lp) Lpid {
	id := len(r.lps)
	if r.lpCnt == id {
		r.lps = append(r.lps, lpEntry{})
	} else {
		for i, _ := range r.lps {
			if r.lps[i].lp == nil {
				id = i
				break
			}
		}
	}
	r.lpCnt += 1

	lpe := &r.lps[id]
	lpe.createTick = s.CurrentTick()
	lpe.runTick = lpe.createTick + 1
	lpe.lp = lp
	lpe.msgs = nil
	return Lpid{r.id, id, lpe.createTick}
}

func (r *Runner) removeLp(idx int) {
	r.lps[idx].runTick = math.MaxInt64
	r.lps[idx].lp = nil
	r.lps[idx].msgs = nil
	r.lpCnt--
}

func (r *Runner) RemoveLp(lpid Lpid) {
	r.removeLp(lpid.Id)
}

func (r *Runner) getLp(lpid Lpid) (Lp, error) {
	if r.id != lpid.RunnerId {
		return nil, fmt.Errorf("Mismatching runner id.")
	}
	if lpid.Id < 0 || lpid.Id >= len(r.lps) {
		return nil, fmt.Errorf("Lpid.Id out of range.")
	}

	lpe := &r.lps[lpid.Id]
	if lpid.CreateTick != lpe.createTick {
		return nil, fmt.Errorf("Mismatching create tick.")
	}
	return lpe.lp, nil
}

func (r *Runner) UpdateAgg(aggId int, ival int64, fval float64) {
	r.aggs[aggId].Update(ival, fval)
}

func (r *Runner) runLoop(s *Sim) {
	for {
		tick := <-r.ch
		if tick < 0 {
			s.ch <- tick
			return
		}

		hasMsg := false
		// First deliever message to each lp
		for i, _ := range r.msgs {
			for _, msg := range r.msgs[i] {
				idx := msg.dst.Id
				lpe := &r.lps[idx]
				if lpe.lp != nil && lpe.createTick <= msg.dst.CreateTick {
					lpe.msgs = append(lpe.msgs, msg.msg)
					hasMsg = true
				}
			}
			r.msgs[i] = nil
		}

		if !hasMsg && r.nextTick > tick {
			// Nothing to run -- just skip to next loop
			s.ch <- r.nextTick
		} else {
			// Next run LPs
			r.nextTick = math.MaxInt64
			for i, _ := range r.lps {
				lpe := &r.lps[i]
				if lpe.lp != nil {
					if len(lpe.msgs) > 0 || lpe.runTick <= tick {
						inc := lpe.lp.Run(s, r, lpe.msgs)
						if inc == -1 {
							r.removeLp(i)
						} else if inc > 0 {
							lpe.runTick = tick + inc
						} else {
							lpe.runTick = math.MaxInt64
						}
					}

					if r.nextTick > lpe.runTick {
						r.nextTick = lpe.runTick
					}
				}
				lpe.msgs = nil
			}
			s.ch <- r.nextTick
		}
	}
}
