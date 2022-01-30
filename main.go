package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// first of all we can't do shit if we don't get into raw mode
// now, before getting into raw mode we need to have a way to
// escape or else we'll be stuck in raw mode and will hae to kill
// the terminal in order to break out

func clearScreen() {
	codes := []string{
		"\x1b[2J",
		"\x1b[H", // CUP
	}
	for _, code := range codes {
		fmt.Fprint(os.Stdout, code)
	}
}

var stdinFd int
var state *term.State
var initError error

func shutDown(stdinFd int, state *term.State) {
	term.Restore(stdinFd, state)
	clearScreen()
}

func init() {
	clearScreen()
	stdinFd = int(os.Stdout.Fd())
	state, initError = term.MakeRaw(stdinFd)
}

func main() {
	if initError != nil {
		fmt.Fprintln(os.Stdout, "err:", initError)
		return
	}
	defer shutDown(stdinFd, state)

	//var term unix.Termios
	//c, r, err := term.GetSize(stdinFd)
	rd := bufio.NewReader(os.Stdin)

OUT:
	for {
		b, err := rd.ReadByte()
		switch err {
		case nil: // no error, process input
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
