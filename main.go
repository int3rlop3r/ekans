package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

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

func shutDown(stdinFd int, state *term.State) {
	term.Restore(stdinFd, state)
	clearScreen(mEnd)
}

func init() {
	clearScreen(mStart)
	stdinFd = int(os.Stdout.Fd())
	state, initError = term.MakeRaw(stdinFd)
}

func processKeyPress() {
OUT:
	for {
		buf := make([]byte, 1)
		_, err := os.Stdin.Read(buf) // dirty way to read from stdin
		switch err {
		case nil: // no error, process input
			b := buf[0]
			fmt.Print(b) // instead of printing we'll have to return events
			if b == 4 || b == 3 {
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
	fmt.Printf("%s\r\n", bytes.Join(buf, []byte("\r\n")))
}

func main() {
	if initError != nil {
		fmt.Fprintln(os.Stderr, "err:", initError)
		return
	}
	defer shutDown(stdinFd, state)

	// we'll use this later to create a buffer so that we can
	// buffer changes and then write them to the screen
	//var term unix.Termios
	c, r, err := term.GetSize(stdinFd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "couldn't get screen size, err:", err)
		return
	}
	fmt.Printf("rows: %d, cols: %d\r\n", r, c)
	buf := makeBuff(r, c)
	buf[10][50] = 'A'
	display(buf)
	time.Sleep(20 * time.Second)
}
