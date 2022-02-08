package main

import (
	"errors"
	"fmt"
	"log"
	"os"

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
var f *os.File

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

	var fErr error
	f, fErr = os.OpenFile("/tmp/snake.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if fErr != nil {
		log.Fatalf("error opening file: %v", fErr)
	}

	log.SetOutput(f)
}

func processKeyPress() (chan error, chan byte) {
	out := make(chan error)
	keyChan := make(chan byte)
	go func() {
		for {
			buf := make([]byte, 1)
			_, err := os.Stdin.Read(buf) // dirty way to read from stdin
			if buf[0] == 3 || buf[0] == 4 {
				out <- errors.New("player quit the game")
			} else if err != nil {
				out <- err
			} else {
				keyChan <- buf[0]
			}
		}
	}()
	return out, keyChan
}

func main() {
	if initError != nil {
		fmt.Fprintln(os.Stderr, "err:", initError)
		return
	}
	defer shutDown(stdinFd, state, exitError)
	defer f.Close()

	c, r, err := term.GetSize(stdinFd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "couldn't get screen size, err:", err)
		return
	}
	exitChan, keyChan := processKeyPress()
	snake := NewSnake()
	game := NewGame(snake, r, c, keyChan)
	game.Start()

OUT:
	for _ = range game.Tick() {
		select {
		case msg := <-exitChan:
			exitError.SetError(msg)
			close(exitChan)
			close(keyChan)
			game.Stop()
			// NOTE: now that we're calling game.Stop()
			// can we get rid of the break statement?
			break OUT
		default:
		}
		game.Refresh()
	}
}
