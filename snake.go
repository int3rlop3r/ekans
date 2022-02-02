package main

const (
	Up = iota
	Down
	Left
	Right
)

type body [2]int

func (s *body) Move() {
	s[1]++
}

type Snake struct {
	Body *[]body
}

func (s *Snake) Move(d *Display) {
	for i := range *s.Body {
		(*s.Body)[i].Move()
		x := (*s.Body)[i]
		d.Plot(x[0], x[1])
	}
}

func NewSnake() *Snake {
	var b []body
	for i := 0; i < 30; i++ {
		b = append(b, [2]int{0, i})
	}
	return &Snake{&b}
}
