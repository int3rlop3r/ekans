package main

import (
	"bytes"
	"fmt"
)

type Display struct {
	buf  [][]byte
	size [2]int
}

func (d *Display) Refresh() {
	for i := range d.buf {
		for j := range d.buf[i] {
			d.buf[i][j] = ' '
		}
	}
}

func (d *Display) Flush() {
	fmt.Printf("\x1b[H\x1b[0J%s\r\n", bytes.Join(d.buf, []byte("\r\n")))
}

func (d *Display) Plot(r, c int) {
	d.buf[r][c] = 'S'
}

func NewDisplay(r, c int) *Display {
	buf := makeBuf(r, c)
	return &Display{buf, [2]int{r, c}}
}

func makeBuf(r, c int) [][]byte {
	buf := make([][]byte, r-1)
	for i := range buf {
		buf[i] = make([]byte, c)
		for j := range buf[i] {
			buf[i][j] = ' ' // fill the buffer with spaces
		}
	}
	return buf
}
