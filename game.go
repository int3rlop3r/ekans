package main

import (
	"bytes"
	"fmt"
	"math/rand"
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

type Game struct {
	buf   [][]byte
	bsize [2]int
	keyCh chan byte
	snake *Snake
	food  [2]int

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
	return !((0 < head[0] && head[0] < g.bsize[0]) && (0 < head[1] && head[1] < g.bsize[1]))
}

func (g *Game) validatePos() {
	if g.touchedBorder() {
		g.snakeIsSafe = false
	} else {
		g.snakeIsSafe = !g.snake.BitSelf()
	}
}

func (g *Game) genFood() {
	pos := g.snake.Head()
	if pos == g.food { // or food is in default location
		x := 0
		for x == 0 {
			x = rand.Intn(g.bsize[0] - 1)
		}
		y := 0
		for y == 0 {
			y = rand.Intn(g.bsize[1] - 1)
		}
		g.food = [2]int{x, y}
	}
}

func (g *Game) Refresh() {
	g.eraseSnake()
	g.plotFood()
	if g.snakeIsSafe {
		g.snake.Move()
		g.genFood()
		g.plotSnake()
		g.snake.TransDir()
		g.validatePos()
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

func NewGame(snake *Snake, r, c int, keyCh chan byte) *Game {
	br, bc, buf := makeBuf(r, c)
	food := [2]int{19, 22}
	bsize := [2]int{br, bc}
	return &Game{buf, bsize, keyCh, snake, food, true}
}

func makeBuf(r, c int) (int, int, [][]byte) {
	buf := make([][]byte, r-1)
	br := r - 3
	bc := c - 1
	for i := range buf {
		buf[i] = make([]byte, c)
		for j := range buf[i] {
			if i == 0 || i == br {
				buf[i][j] = '-' // fill the buffer with spaces
			} else if (j == 0 || j == bc) && i < br {
				buf[i][j] = '|' // fill the buffer with spaces
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

type Snake struct {
	Body *[]cell
}

func (s *Snake) Move() {
	for i := range *s.Body {
		(*s.Body)[i].Move()
	}
}

func (s *Snake) Head() [2]int {
	return (*s.Body)[0].pos
}

func (s *Snake) TransDir() {
	for i := len(*s.Body) - 1; i > 0; i-- {
		(*s.Body)[i].dir = (*s.Body)[i-1].dir
	}
}

func (s *Snake) BitSelf() bool {
	head := (*s.Body)[0].pos
	for i := 1; i < len(*s.Body)-1; i++ {
		if head[0] == (*s.Body)[i].pos[0] && head[1] == (*s.Body)[i].pos[1] {
			return true
		}
	}
	return false
}

func (s *Snake) ChangeDir(key byte) {
	curDir := (*s.Body)[0].dir
	var dir int
	switch {
	case key == KpUp && curDir != Down:
		dir = Up
	case key == KpDown && curDir != Up:
		dir = Down
	case key == KpLeft && curDir != Right:
		dir = Left
	case key == KpRight && curDir != Left:
		dir = Right
	default:
		return
	}
	(*s.Body)[0].dir = dir // change dir
}

func NewSnake() *Snake {
	var b []cell
	for i := 30; i > 5; i-- {
		b = append(b, cell{dir: Right, pos: [2]int{1, i}})
	}
	return &Snake{&b}
}
