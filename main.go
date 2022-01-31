package main

import (
	"bytes"
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

type KeyPress struct {
	Code byte
	Err  error
}

func processKeyPress() chan *KeyPress {
	out := make(chan *KeyPress)
	go func() {
		for {
			buf := make([]byte, 1)
			_, err := os.Stdin.Read(buf) // dirty way to read from stdin
			out <- &KeyPress{buf[0], err}
			//switch err {
			//case nil: // no error, process input
			//b := buf[0]
			////fmt.Print(b) // instead of printing we'll have to return events
			//if b == 4 || b == 3 {
			//fmt.Print("qutting!\r\n")
			//break OUT
			//}
			//out <- b
			//case io.EOF:
			//break
			//default:
			//fmt.Print("err: %s\r\n", err)
			//break
			//}
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
	defer shutDown(stdinFd, state)

	// we'll use this later to create a buffer so that we can
	// buffer changes and then write them to the screen
	//var term unix.Termios
	c, r, err := term.GetSize(stdinFd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "couldn't get screen size, err:", err)
		return
	}
	//fmt.Printf("rows: %d, cols: %d\r\n", r, c)
	buf := makeBuff(r, c)
	kpChan := processKeyPress()
	snake := NewSnake()
OUT:
	for _ = range time.Tick(1 * time.Second) {
		select {
		case ev := <-kpChan:
			if ev.Err != nil {
				fmt.Fprint(os.Stderr, "key press error:", ev.Err)
			}
			key := ev.Code
			if key == 4 || key == 3 {
				fmt.Print("qutting!\r\n")
				break OUT
			}
			// we'll probably have to flush here too
			// but we'll worry about it later!!
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
