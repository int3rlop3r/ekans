package main

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

type cell struct {
	pos [2]int
	dir int
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
	body    *[]cell
	keyChan chan byte
}

func (s *Snake) Move(d *Display) {
	for i := range *s.body {
		(*s.body)[i].Move()
		x := (*s.body)[i].pos
		d.Plot(x[0], x[1], '*')
	}
	for i := len(*s.body) - 1; i > 0; i-- {
		(*s.body)[i].dir = (*s.body)[i-1].dir
	}
}

func (s *Snake) Bind() {
	go func() {
		for key := range s.keyChan {
			if !s.validKP(key) {
				continue
			}
			s.changeDir(key)
		}
	}()
}

func (s *Snake) validKP(key byte) bool {
	return key == KpUp || key == KpDown ||
		key == KpLeft || key == KpRight
}

func (s *Snake) changeDir(key byte) {
	var dir int
	switch key {
	case KpUp:
		dir = Up
	case KpDown:
		dir = Down
	case KpLeft:
		dir = Left
	case KpRight:
		dir = Right
	}
	(*s.body)[0].dir = dir
}

func NewSnake(keyChan chan byte) *Snake {
	var b []cell
	for i := 30; i > 0; i-- {
		b = append(b, cell{dir: Right, pos: [2]int{0, i}})
	}
	return &Snake{&b, keyChan}
}
