package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// first of all we can't do shit if we don't get into raw mode
// now, before getting into raw mode we need to have a way to
// escape or else we'll be stuck in raw mode and will hae to kill
// the terminal in order to break out

const (
	ClrScr   = "\x1b[2J"
	CurPosUp = "\x1b[H" // CUP
)

func clearJunk() {
	fmt.Fprint(os.Stdout, ClrScr)
	fmt.Fprint(os.Stdout, CurPosUp)
}

func main() {
	clearJunk()
	//var term unix.Termios
	//fmt.Println(term)
	stdinFd := int(os.Stdout.Fd())
	state, err := term.MakeRaw(stdinFd)
	defer term.Restore(stdinFd, state)
	c, r, err := term.GetSize(stdinFd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "err:", err, c)
	}
	//fmt.Println(c, r)
	for i := 0; i < r; i++ {
		fmt.Print("~\r\n")
	}
}
