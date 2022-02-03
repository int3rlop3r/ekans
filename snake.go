package main

const (
	Up = iota
	Down
	Left
	Right
)

type body struct {
	pos [2]int
	dir int
}

func (s *body) Move() {
	switch s.dir {
	case Right:
		s.pos[1]++
	}
}

type Snake struct {
	Body *[]body
}

func (s *Snake) Move(d *Display) {
	for i := range *s.Body {
		(*s.Body)[i].Move()
		x := (*s.Body)[i].pos
		d.Plot(x[0], x[1])
	}
}

func NewSnake() *Snake {
	var b []body
	for i := 0; i < 30; i++ {
		b = append(b, body{dir: Right, pos: [2]int{0, i}})
	}
	return &Snake{&b}
}
