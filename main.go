package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

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
