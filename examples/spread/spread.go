package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/fengttt/gosim"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

const (
	// Universe size NxN
	N = 1024
	// Infectious for 10 steps.
	PERIOD = 10
	// 16 Runners
	NRunner = 16
	// Each PL will update a 16x16 cells
	NCELL = 16
)

type universe struct {
	rate [5]float64
	data [2][N][N]int
}

func (u *universe) init(n int, r float64) {
	for i := 0; i < n; i++ {
		idx := rand.Intn(N * N)
		u.data[0][idx/N][idx%N] = 1
		u.data[1][idx/N][idx%N] = 1
	}

	u.rate[0] = 0
	u.rate[1] = r
	u.rate[2] = 1 - math.Pow(1-r, 2)
	u.rate[3] = 1 - math.Pow(1-r, 3)
	u.rate[4] = 1 - math.Pow(1-r, 4)
}

func (u *universe) current(tick int64) *[N][N]int {
	return &u.data[tick%2]
}

func (u *universe) next(tick int64) *[N][N]int {
	return &u.data[1-tick%2]
}

var u universe
var sim *gosim.Sim
var totalAgg int

// A patch of NCELLxNCELL cells
type cells struct {
	// Lower left corner
	x, y int
}

func testLive(data *[N][N]int, x, y int, tick int) int {
	if x < 0 || y < 0 || x >= N || y >= N {
		return 0
	}
	if data[x][y] > 0 && data[x][y] > tick-PERIOD {
		return 1
	}
	return 0
}

func cntNeighbours(b *[N][N]int, x, y int, tick int) int {
	return testLive(b, x-1, y, tick) + testLive(b, x+1, y, tick) +
		testLive(b, x, y-1, tick) + testLive(b, x, y+1, tick)
}

func (c *cells) Run(s *gosim.Sim, r *gosim.Runner, msgs []interface{}) int64 {
	tick := s.CurrentTick()
	curr := u.current(tick)
	next := u.next(tick)
	for i := 0; i < NCELL; i++ {
		for j := 0; j < NCELL; j++ {
			xx := c.x + i
			yy := c.y + j
			if next[xx][yy] == 0 {
				cnt := cntNeighbours(curr, xx, yy, int(tick))
				if cnt > 0 {
					rr := r.Rand().Float64()
					if rr < u.rate[cnt] {
						next[xx][yy] = int(tick)
						curr[xx][yy] = int(tick)
						r.UpdateAgg(totalAgg, 1, 1)
					}
				}
			}
		}
	}
	return 1
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "GoL",
		Bounds: pixel.R(0, 0, N, N),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	basicTxt := text.New(pixel.V(N/2, 10), basicAtlas)
	basicTxt.Color = colornames.Blue

	imd := imdraw.New(nil)
	imd.Clear()
	imd.Color = pixel.RGB(1, 0, 0)
	var drawn [N][N]bool

	for !win.Closed() {
		tick := sim.CurrentTick()
		b := u.current(tick)

		for i := 0; i < N; i++ {
			for j := 0; j < N; j++ {
				if b[i][j] > 0 && !drawn[i][j] {
					xx := float64(i)
					yy := float64(j)
					imd.Push(pixel.V(xx, yy))
					imd.Push(pixel.V(xx+1, yy+1))
					imd.Rectangle(0)
					drawn[i][j] = true
				}
			}
		}

		win.Clear(colornames.White)
		imd.Draw(win)

		basicTxt.Clear()
		nci, _ := sim.ReadAgg(totalAgg)
		fmt.Fprintf(basicTxt, "tick %d: total infected %d", tick, nci)
		basicTxt.Draw(win, pixel.IM)

		win.Update()
		sim.RunSteps(1)
	}
}

func main() {
	var flagN = flag.Int("n", 100, "number of initial infected")
	var flagR = flag.Float64("r", 0.1, "rate of infection in one period")
	flag.Parse()
	u.init(*flagN, *flagR)

	sim = gosim.New(NRunner)
	for x := 0; x < N; x += NCELL {
		for y := 0; y < N; y += NCELL {
			cs := new(cells)
			cs.x = x
			cs.y = y
			sim.AddLp(((x+y)/NCELL)%NRunner, cs)
		}
	}
	totalAgg = sim.CreateSum()

	sim.Start()
	pixelgl.Run(run)
}
