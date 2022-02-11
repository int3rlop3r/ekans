package main

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"
)

const (
	Up = iota
	Down
	Left
	Right
	KpUp    = 0x41
	KpDown  = 0x42
	KpLeft  = 0x44
	KpRight = 0x43
)

type GameTicker struct {
	d      time.Duration
	ticker *time.Ticker
}

func (t *GameTicker) GetChan() <-chan time.Time {
	return t.ticker.C
}

func (t *GameTicker) Stop() {
	t.ticker.Stop()
}

func (t *GameTicker) Dur() string {
	return t.d.String()
}

func (t *GameTicker) Incr() {
	t.d = time.Duration(float64(t.d) - (float64(t.d) * math.Pow(0.5, 5.0))) //  * time.Microsecond
	t.ticker.Reset(t.d)
}

type Game struct {
	buf    [][]byte
	bsize  [2]int
	keyCh  chan byte
	snake  *Snake
	food   [2]int
	ticker *GameTicker
	score  int

	// becomes false if snake touches
	// borders or bites itself
	snakeIsSafe bool
}

func (g *Game) eraseSnake() {
	for _, cell := range *g.snake.Body {
		r, c := cell.GetPos()
		g.buf[r][c] = ' '
	}
}

func (g *Game) plotSnake() {
	for _, cell := range *g.snake.Body {
		r, c := cell.GetPos()
		g.buf[r][c] = '*'
	}
}

func (g *Game) plotFood() {
	g.buf[g.food[0]][g.food[1]] = '$'
}

func (g *Game) touchedBorder() bool {
	head := g.snake.Head()
	return !((0 < head[0] && head[0] < g.bsize[0]) &&
		(0 < head[1] && head[1] < g.bsize[1]))
}

func (g *Game) genFood() {
	pos := g.snake.Head()
	if pos == g.food { // or food is in default location
		// NOTE: since this is a bit slow we can pre-compute
		// values and then consume them as they're used.
		// We can use a buffered channel for this.
		x := rand.Intn(g.bsize[0] - 1)
		y := rand.Intn(g.bsize[1] - 1)
		if x == 0 {
			x++
		}
		if y == 0 {
			y++
		}
		g.food = [2]int{x, y}
		g.snake.Grow()
		g.IncrScoreAndTicker()
		g.DisplayScore()
	}
}

func (g *Game) Refresh() {
	g.eraseSnake()
	g.plotFood()
	if g.snakeIsSafe {
		if err := g.snake.Move(); err != nil {
			g.snakeIsSafe = false
		}
		g.genFood()
		g.plotSnake()
		if g.touchedBorder() {
			g.snakeIsSafe = false
		}
	} else {
		g.plotSnake()
		g.plotGameOver()
	}
	g.flush()
}

func (g *Game) plotGameOver() {
	msg := []byte(" You lose! Press Ctrl+C to quit. ")
	r := 5
	c := 30
	// NOTE: this will make the game crash if the screen is too small
	for i, ch := range msg {
		g.buf[r][c+i] = ch
	}
}

func (g *Game) flush() {
	fmt.Printf("\x1b[H\x1b[0J%s\r\n", bytes.Join(g.buf, []byte("\r\n")))
}

func (g *Game) Plot(r, c int, ch byte) {
	g.buf[r][c] = ch
}

func (g *Game) Start() {
	g.DisplayScore()
	go func() {
		for key := range g.keyCh {
			if !g.validKP(key) {
				continue
			}
			g.snake.ChangeDir(key)
		}
	}()
}

func (g *Game) validKP(key byte) bool {
	return key == KpUp || key == KpDown ||
		key == KpLeft || key == KpRight
}

func (g *Game) Tick() <-chan time.Time {
	return g.ticker.GetChan()
}

func (g *Game) Stop() {
	g.ticker.Stop()
}

func (g *Game) IncrScoreAndTicker() {
	g.ticker.Incr()
	g.score += 50
	//log.Println("changed duration to:", g.ticker.Dur())
}

func (g *Game) DisplayScore() {
	strScore := []byte("Score: ")
	strScore = append(strScore, []byte(strconv.Itoa(g.score))...)
	for i, ch := range strScore {
		g.buf[g.bsize[0]+1][i] = ch
	}
}

func NewGame(snake *Snake, r, c int, keyCh chan byte) *Game {
	br, bc, buf := makeBuf(r, c)
	food := [2]int{19, 22}
	bsize := [2]int{br, bc}
	tickDur := 120 * time.Millisecond
	ticker := &GameTicker{tickDur, time.NewTicker(tickDur)}
	return &Game{buf, bsize, keyCh, snake, food, ticker, 0, true}
}

func makeBuf(r, c int) (int, int, [][]byte) {
	buf := make([][]byte, r-1)
	br := r - 3
	bc := c - 1
	for i := range buf {
		buf[i] = make([]byte, c)
		for j := range buf[i] {
			if i == 0 || i == br {
				buf[i][j] = '-' // top and bottom border
			} else if (j == 0 || j == bc) && i < br {
				buf[i][j] = '|' // left and right border
			} else {
				buf[i][j] = ' ' // fill the buffer with spaces
			}
		}
	}
	return br, bc, buf
}

type cell struct {
	pos [2]int
	dir int
}

func (s *cell) GetPos() (int, int) {
	return s.pos[0], s.pos[1]
}

func (s *cell) Move() {
	switch s.dir {
	case Up:
		s.pos[0]--
	case Down:
		s.pos[0]++
	case Left:
		s.pos[1]--
	case Right:
		s.pos[1]++
	}
}

func (s *cell) GetPrevPos() [2]int {
	switch s.dir {
	case Up:
		return [2]int{s.pos[0] + 1, s.pos[1]}
	case Down:
		return [2]int{s.pos[0] - 1, s.pos[1]}
	case Left:
		return [2]int{s.pos[0], s.pos[1] + 1}
	case Right:
		return [2]int{s.pos[0], s.pos[1] - 1}
	}
	return [2]int{}
}

type Snake struct {
	Body *[]cell
}

func (s *Snake) Move() error {
	var err error
	prevDir := (*s.Body)[0].dir
	for i := range *s.Body {
		(*s.Body)[i].Move()
		if i == 0 {
			continue
		}

		if (*s.Body)[i].pos == (*s.Body)[0].pos {
			err = errors.New("snake bit self")
		}

		// transfer direction
		tmp := (*s.Body)[i].dir
		(*s.Body)[i].dir = prevDir
		prevDir = tmp
	}
	return err // complete movement then report
}

func (s *Snake) Head() [2]int {
	return (*s.Body)[0].pos
}

func (s *Snake) ChangeDir(key byte) {
	prevDir := (*s.Body)[1].dir
	var dir int
	switch {
	case key == KpUp && prevDir != Down:
		dir = Up
	case key == KpDown && prevDir != Up:
		dir = Down
	case key == KpLeft && prevDir != Right:
		dir = Left
	case key == KpRight && prevDir != Left:
		dir = Right
	default:
		return
	}
	(*s.Body)[0].dir = dir
}

func (s *Snake) Grow() {
	last := (*s.Body)[len((*s.Body))-1]
	(*s.Body) = append((*s.Body), cell{dir: last.dir, pos: last.GetPrevPos()})
}

func NewSnake() *Snake {
	var b []cell
	for i := 20; i > 5; i-- {
		b = append(b, cell{dir: Right, pos: [2]int{1, i}})
	}
	return &Snake{&b}
}
