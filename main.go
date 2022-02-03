package main

import (
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
			if buf[0] == 3 || buf[0] == 4 {
				out <- errors.New("player quit the game")
			} else if err != nil {
				out <- err
			}
		}
	}()
	return out
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
	display := NewDisplay(r, c)
	exitGame := processKeyPress()
	snake := NewSnake()

OUT:
	for _ = range time.Tick(100 * time.Millisecond) {
		select {
		case msg := <-exitGame:
			// create a new type 'key press', it will return the
			// direction that was pressed - means validation will be taken care of
			// {dir, err}, on error quit the game other wise process the event
			exitError.SetError(msg)
			break OUT
		default:
			// this will go out of the default block and into the main 'for'
			// so that events can be passed as value or nil (or -1)
			// WE STILL NEED TO FIGURE OUT HOW TO TURN, AND PASS TURNING
			// INFORMATION TO THE REST OF THE BODY
			// ALSO, DRAW BORDERS SO THAT PEOPLE KNOW WHERE THE SCREEN ENDS!!
			display.Refresh()
			snake.Move(display)
			display.Flush()
		}
	}
}
