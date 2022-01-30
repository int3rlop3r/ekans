package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

const (
	mStart = iota
	mEnd
	mBoth
)

type Inst struct {
	Mode int
	Code string
}

func clearScreen(mode int) {
	instrs := []Inst{
		{mBoth, "\x1b[2J"},    // clear the screen
		{mBoth, "\x1b[H"},     // CUP - get the cursor UP (top left)
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

func shutDown(stdinFd int, state *term.State) {
	term.Restore(stdinFd, state)
	clearScreen(mEnd)
}

func init() {
	clearScreen(mStart)
	stdinFd = int(os.Stdout.Fd())
	state, initError = term.MakeRaw(stdinFd)
}

func main() {
	if initError != nil {
		fmt.Fprintln(os.Stdout, "err:", initError)
		return
	}
	defer shutDown(stdinFd, state)

	// we'll use this later to create a buffer so that we can
	// buffer changes and then write them to the screen
	//var term unix.Termios
	//c, r, err := term.GetSize(stdinFd)

OUT:
	for {
		buf := make([]byte, 1)
		_, err := os.Stdin.Read(buf) // dirty way to read from stdin
		switch err {
		case nil: // no error, process input
			b := buf[0]
			fmt.Print(b)
			if b == 4 {
				fmt.Print("qutting!\r\n")
				break OUT
			}
		case io.EOF:
			break
		default:
			fmt.Print("err: %s\r\n", err)
			break
		}
	}
}
