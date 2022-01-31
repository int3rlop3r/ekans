package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

const (
	mStart = iota
	mEnd
	mBoth
)

type ExitError struct {
	msg string
}

func (e *ExitError) SetError(msg error) {
	e.msg = msg.Error()
}

func (e *ExitError) Error() string {
	return e.msg
}

type Inst struct {
	Mode int
	Code string
}

func clearScreen(mode int) {
	instrs := []Inst{
		{mBoth, "\x1b[2J"}, // clear the screen
		{mBoth, "\x1b[H"},  // CUP - get the cursor UP (top left)
		//{mStart, "\x1b[2J"},   // clear the screen
		//{mStart, "\x1b[H"},    // CUP - get the cursor UP (top left)
		{mStart, "\x1b[?25l"}, // hide the cursor
		{mEnd, "\x1b[?25h"},   // display the cursor
	}

	for _, inst := range instrs {
		currMode := inst.Mode
		if currMode == mBoth || currMode == mode {
			fmt.Fprint(os.Stdout, inst.Code)
		}
	}
}

var stdinFd int
var state *term.State
var initError error
var exitError *ExitError

func shutDown(stdinFd int, state *term.State, err error) {
	term.Restore(stdinFd, state)
	clearScreen(mEnd)
	fmt.Println(err)
}

func init() {
	clearScreen(mStart)
	stdinFd = int(os.Stdout.Fd())
	state, initError = term.MakeRaw(stdinFd)
	exitError = new(ExitError)
}

func processKeyPress() chan error {
	out := make(chan error)
	go func() {
		for {
			buf := make([]byte, 1)
			_, err := os.Stdin.Read(buf) // dirty way to read from stdin
			//out <- &KeyPress{buf[0], err}
			//if buf[0] == 3 || buf[0] == 4 || err != nil {
			if buf[0] == 3 || buf[0] == 4 {
				//fmt.Println(err) // NOTE: this never gets printed
				out <- errors.New("player quit the game")
			} else if err != nil {
				out <- err
			}
		}
	}()
	return out
}

func makeBuff(r, c int) [][]byte {
	buf := make([][]byte, r-1)
	for i := range buf {
		buf[i] = make([]byte, c)
		for j := range buf[i] {
			buf[i][j] = ' ' // fill the buffer with spaces
		}
	}
	return buf
}

func display(buf [][]byte) {
	fmt.Printf("\x1b[H\x1b[0J%s\r\n", bytes.Join(buf, []byte("\r\n")))
}

func main() {
	if initError != nil {
		fmt.Fprintln(os.Stderr, "err:", initError)
		return
	}
	defer shutDown(stdinFd, state, exitError)

	c, r, err := term.GetSize(stdinFd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "couldn't get screen size, err:", err)
		return
	}
	//fmt.Printf("rows: %d, cols: %d\r\n", r, c)
	buf := makeBuff(r, c)
	exitGame := processKeyPress()
	snake := NewSnake()
OUT:
	for _ = range time.Tick(1 * time.Second) {
		select {
		case msg := <-exitGame:
			exitError.SetError(msg)
			break OUT
		default:
			snake.Move()
			plot(buf, snake)
			display(buf)
		}
	}
}

func plot(buf [][]byte, snake *Snake) {
	buf[(*snake)[0]][(*snake)[1]] = 'S'
}
