package gosim

import (
	"math"
)

type Agg interface {
	Update(ival int64, fval float64)
	Get() (int64, float64)
	Gather(a Agg)
	Reset()
}

type cntSum struct {
	cnt int64
	sum float64
}

func newSum() *cntSum {
	var c cntSum
	c.Reset()
	return &c
}

func (c *cntSum) Update(ival int64, fval float64) {
	c.cnt += ival
	c.sum += fval
}

func (c *cntSum) Gather(a Agg) {
	i, f := a.Get()
	c.cnt += i
	c.sum += f
}

func (c *cntSum) Get() (int64, float64) {
	return c.cnt, c.sum
}

func (c *cntSum) Reset() {
	c.cnt = 0
	c.sum = 0
}

type cntMin struct {
	cnt int64
	min float64
}

func newMin() *cntMin {
	var c cntMin
	c.Reset()
	return &c
}

func (c *cntMin) Update(ival int64, fval float64) {
	c.cnt += ival
	c.min = math.Min(c.min, fval)
}

func (c *cntMin) Get() (int64, float64) {
	return c.cnt, c.min
}

func (c *cntMin) Gather(a Agg) {
	i, f := a.Get()
	c.cnt += i
	c.min += math.Min(c.min, f)
}

func (c *cntMin) Reset() {
	c.cnt = 0
	c.min = math.Inf(1)
}

type cntMax struct {
	cnt int64
	max float64
}

func newMax() *cntMax {
	var c cntMax
	c.Reset()
	return &c
}

func (c *cntMax) Update(ival int64, fval float64) {
	c.cnt += ival
	c.max = math.Max(c.max, fval)
}

func (c *cntMax) Gather(a Agg) {
	i, f := a.Get()
	c.cnt += i
	c.max += math.Max(c.max, f)
}

func (c *cntMax) Get() (int64, float64) {
	return c.cnt, c.max
}

func (c *cntMax) Reset() {
	c.cnt = 0
	c.max = math.Inf(-1)
}
