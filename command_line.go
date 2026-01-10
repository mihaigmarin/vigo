package main

import (
	"unicode"
)

// Command line struct. Commands are stored here.
type cmdl struct {
	buf []rune
	c   cursor
}

// Initialize command line. The parameter "h" is in fact the height where
// the command line writes text. It will always be "h-1", where "h" is the
// height calculated from the "tcell.Screen".
func (cl *cmdl) init(h int) {
	cl.buf = []rune{}
	cl.c.x = 0
	cl.c.y = h - 1
	cl.c.offset = 0
}

// Put rune to command buffer.
func (cl *cmdl) put(r rune) {
	if !unicode.IsControl(r) {
		cl.buf = append(cl.buf, r)
		cl.c.x++
	}
}

// Handle backspace press in insert mode. When triggered this function
// deletes the rune after the cursor position in the command line buffer.
func (cl *cmdl) backspace() {
	if cl.c.x > 0 {
		cl.buf = cl.buf[:cl.c.x-1]
		cl.c.x--
	}
}

// Reset command buffer.
func (cl *cmdl) reset() {
	cl.buf = []rune{}
	cl.c.x = 0
}
