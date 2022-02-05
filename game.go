package main

import (
	"bytes"
	"fmt"
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
	size  [2]int
	keyCh chan byte
	snake *Snake
}

func (d *Game) Refresh() {
	for _, cell := range *d.snake.Body {
		r, c := cell.GetPos()
		d.buf[r][c] = ' '
	}
}

func (d *Game) Flush() {
	fmt.Printf("\x1b[H\x1b[0J%s\r\n", bytes.Join(d.buf, []byte("\r\n")))
}

func (d *Game) Plot(r, c int, ch byte) {
	d.buf[r][c] = ch
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
	buf := makeBuf(r, c)
	return &Game{buf, [2]int{r, c}, keyCh, snake}
}

func makeBuf(r, c int) [][]byte {
	buf := make([][]byte, r-1)
	for i := range buf {
		buf[i] = make([]byte, c)
		for j := range buf[i] {
			buf[i][j] = ' ' // fill the buffer with spaces
		}
	}
	return buf
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

func (s *Snake) Move(d *Game) {
	for i := range *s.Body {
		(*s.Body)[i].Move()
		x := (*s.Body)[i].pos
		d.Plot(x[0], x[1], '*')
	}
	for i := len(*s.Body) - 1; i > 0; i-- {
		(*s.Body)[i].dir = (*s.Body)[i-1].dir
	}
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
	for i := 30; i > 0; i-- {
		b = append(b, cell{dir: Right, pos: [2]int{0, i}})
	}
	return &Snake{&b}
}
