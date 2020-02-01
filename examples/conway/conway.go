package main

import (
	"fmt"
	"os"

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
	N = 256
	// Cell width for drawing
	CW = 2
	// 16 Runners
	NRunner = 16
	// Each PL will update a 16x16 cells
	NCELL = 16
)

type universe struct {
	data [2][N][N]bool
}

func (u *universe) init(s string) {
	for i := 0; i < 2; i++ {
		if s == "block" {
			pts := [4][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}}
			for n := 0; n < 4; n++ {
				u.data[i][N/2+pts[n][0]][N/2+pts[n][1]] = true
			}
		} else if s == "blinker" {
			pts := [3][2]int{{0, 0}, {0, 1}, {0, 2}}
			for n := 0; n < 3; n++ {
				u.data[i][N/2+pts[n][0]][N/2+pts[n][1]] = true
			}
		} else if s == "beacon" {
			pts := [6][2]int{{0, 2}, {0, 3}, {1, 3}, {2, 0}, {3, 0}, {3, 1}}
			for n := 0; n < 6; n++ {
				u.data[i][N/2+pts[n][0]][N/2+pts[n][1]] = true
			}
		} else if s == "glider" {
			pts := [5][2]int{{0, 0}, {1, 0}, {1, 2}, {2, 0}, {2, 1}}
			for n := 0; n < 5; n++ {
				u.data[i][N/2+pts[n][0]][N/2+pts[n][1]] = true
			}
		} else if s == "r" {
			pts := [5][2]int{{0, 1}, {1, 0}, {1, 1}, {1, 2}, {2, 2}}
			for n := 0; n < 5; n++ {
				u.data[i][N/2+pts[n][0]][N/2+pts[n][1]] = true
			}
		} else if s == "diehard" {
			pts := [7][2]int{{0, 1}, {1, 0}, {1, 1}, {5, 0}, {6, 0}, {6, 2}, {7, 0}}
			for n := 0; n < 7; n++ {
				u.data[i][N/2+pts[n][0]][N/2+pts[n][1]] = true
			}
		} else if s == "inf1" {
			pts := [10][2]int{{0, 0}, {2, 0}, {2, 1}, {4, 2}, {4, 3}, {4, 4}, {6, 3}, {6, 4}, {6, 5}, {7, 4}}
			for n := 0; n < 10; n++ {
				u.data[i][N/2+pts[n][0]][N/2+pts[n][1]] = true
			}
		}
	}
}

func (u *universe) current(tick int64) *[N][N]bool {
	return &u.data[tick%2]
}

func (u *universe) next(tick int64) *[N][N]bool {
	return &u.data[1-tick%2]
}

var u universe
var sim *gosim.Sim
var changeAgg int

// A patch of NCELLxNCELL cells
type cells struct {
	// Lower left corner
	x, y int
}

func testLive(b *[N][N]bool, x, y int) int {
	// Out of board, treat as dead
	if x < 0 || y < 0 || x >= N || y >= N {
		return 0
	}
	if b[x][y] {
		return 1
	}
	return 0
}

func cntNeighbours(b *[N][N]bool, x, y int) int {
	return testLive(b, x-1, y-1) + testLive(b, x-1, y) + testLive(b, x-1, y+1) +
		testLive(b, x, y-1) + 0 + testLive(b, x, y+1) +
		testLive(b, x+1, y-1) + testLive(b, x+1, y) + testLive(b, x+1, y+1)
}

func (c *cells) Run(s *gosim.Sim, r *gosim.Runner, msgs []interface{}) int64 {
	tick := s.CurrentTick()
	curr := u.current(tick)
	next := u.next(tick)
	for i := 0; i < NCELL; i++ {
		for j := 0; j < NCELL; j++ {
			xx := c.x + i
			yy := c.y + j
			cnt := cntNeighbours(curr, xx, yy)
			if cnt == 3 {
				next[xx][yy] = true
			} else if cnt == 2 && curr[xx][yy] {
				next[xx][yy] = true
			} else {
				next[xx][yy] = false
			}

			if next[xx][yy] != curr[xx][yy] {
				r.UpdateAgg(changeAgg, 1, 1)
			}
		}
	}
	return 1
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "GoL",
		Bounds: pixel.R(0, 0, CW*N, CW*N),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	basicTxt := text.New(pixel.V(CW*N/2, 10), basicAtlas)
	basicTxt.Color = colornames.Blue

	loopCnt := 0
	imd := imdraw.New(nil)
	for !win.Closed() {
		imd.Clear()
		imd.Color = pixel.RGB(0, 0, 0)

		tick := sim.CurrentTick()
		b := u.current(tick)
		nLive := 0
		for i := 0; i < N; i++ {
			for j := 0; j < N; j++ {
				if b[i][j] {
					xx := float64(CW * i)
					yy := float64(CW * j)
					imd.Push(pixel.V(xx, yy))
					imd.Push(pixel.V(xx+2, yy+2))
					imd.Rectangle(0)
					nLive++
				}
			}
		}

		win.Clear(colornames.White)
		imd.Draw(win)

		basicTxt.Clear()
		nci, _ := sim.ReadAgg(changeAgg)
		fmt.Fprintf(basicTxt, "tick %d: nLive %d, nChange %d.", tick, nLive, nci)
		basicTxt.Draw(win, pixel.IM)

		win.Update()

		loopCnt += 1
		if loopCnt%10 == 0 {
			sim.RunSteps(1)
		}
	}
}

func main() {
	if len(os.Args) == 1 {
		u.init("glider")
	} else {
		u.init(os.Args[1])
	}

	sim = gosim.New(NRunner)
	for x := 0; x < N; x += NCELL {
		for y := 0; y < N; y += NCELL {
			cs := new(cells)
			cs.x = x
			cs.y = y
			sim.AddLp(((x+y)/NCELL)%NRunner, cs)
		}
	}
	changeAgg = sim.CreateSum()

	sim.Start()
	pixelgl.Run(run)
}
