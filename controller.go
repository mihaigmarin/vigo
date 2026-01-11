package main

import (
	"os"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
)

type controller struct {
    m       *model
    v       *view
	c       cursor
	lastkey rune
}

// Init editor
func (c *controller) init(m *model, v *view) {
    c.m = m
    c.v = v
	c.c.init()
}

// Move editor cursor up.
func (c *controller) up() {
	if c.c.y > 0 {
		c.c.y--
	} else if c.c.offset > 0 {
		c.c.offset--
	}
	line := c.m.lines[c.c.y+c.c.offset]
	limit := len(line) - 1
	if c.m.mode == INSERT {
		limit = len(line)
	}
	if c.c.x > limit {
		if limit < 0 {
			c.c.x = 0
		} else {
			c.c.x = limit
		}
	}
}

// Move editor cursor down.
func (c *controller) down() {
	_, h := c.v.screen.Size()
	if c.c.y+c.c.offset <= len(c.m.lines)-2 {
		if c.c.y < h-2 {
			c.c.y++
		} else {
			c.c.offset++
		}
		line := c.m.lines[c.c.y+c.c.offset]
		limit := len(line) - 1
		if c.m.mode == INSERT {
			limit = len(line)
		}
		if c.c.x > limit {
			if limit < 0 {
				c.c.x = 0
			} else {
				c.c.x = limit
			}
		}
	}
}

// Move editor cursor left.
func (c *controller) left() {
	if c.c.x > 0 {
		c.c.x--
	}
}

// Move editor cursor right.
func (c *controller) right() {
	line := c.m.lines[c.c.y+c.c.offset]
	limit := len(line) - 1
	if c.m.mode == INSERT {
		limit = len(line)
	}
	if c.c.x < limit {
		c.c.x++
	}
}

// Move cursor left until the end of word. Triggered with "b" in normal mode
func (c *controller) leftword() {
	i := c.c.y + c.c.offset

	c.left()
	for c.c.x >= 0 && unicode.IsSpace(c.m.lines[i][c.c.x]) {
		c.left()
	}
	for c.c.x-1 >= 0 {
		if unicode.IsSpace(c.m.lines[i][c.c.x-1]) {
			break
		}
		c.left()
	}
}

// Move cursor right until the end of word. Triggered with "e" in normal mode
func (c *controller) rightword() {
	i := c.c.y + c.c.offset

	c.right()
	for c.c.x < len(c.m.lines[i]) && unicode.IsSpace(c.m.lines[i][c.c.x]) {
		c.right()
	}
	for c.c.x+1 < len(c.m.lines[i]) {
		if unicode.IsSpace(c.m.lines[i][c.c.x+1]) {
			break
		}
		c.right()
	}
}

// put rune to screen
func (c *controller) put(r rune) {
	// put only printable character, avoid control characters such
	// Enter, Backspace, etc. Those are handled separatly in the
	// function 'handleKey'
	if !unicode.IsControl(r) {
		pl := &c.m.lines[c.c.y+c.c.offset]
		*pl = append(*pl, 0)
		copy((*pl)[c.c.x+1:], (*pl)[c.c.x:])
		(*pl)[c.c.x] = r
		c.c.x++
	}
}

// Handle backspace press in insert mode. When triggered this function
// deletes the rune after the cursor position in the editor buffer.
func (c *controller) backspace() {
	if c.c.x > 0 {
		pl := &c.m.lines[c.c.y+c.c.offset]
		*pl = append((*pl)[:c.c.x-1], (*pl)[c.c.x:]...)
		c.c.x--
	} else if c.c.y+c.c.offset > 0 {
		i := c.c.y + c.c.offset
		prevLineLen := len(c.m.lines[i-1])
		c.m.lines[i-1] = append(c.m.lines[i-1], c.m.lines[i]...)
		c.m.lines = append(c.m.lines[:i], c.m.lines[i+1:]...)
		if c.c.y > 0 {
			c.c.y--
		} else if c.c.offset > 0 {
			c.c.offset--
		}
		c.c.x = prevLineLen
	}
}

func (c *controller) deletechar() {
	i := c.c.y + c.c.offset
	pl := &c.m.lines[i]

	if len(*pl) == 0 {
		return
	}

	if len(*pl) == 1 {
		*pl = []rune{}
		c.c.x = 0
		return
	}

	if c.c.x < len(*pl) {
		*pl = append((*pl)[:c.c.x], (*pl)[c.c.x+1:]...)
		if c.c.x >= len(*pl) {
			c.c.x = len(*pl) - 1
		}
	}
}

// Add a new line
func (c *controller) newline() {
	i := c.c.y + c.c.offset + 1
	c.m.lines = append(c.m.lines, nil)
	copy(c.m.lines[i+1:], c.m.lines[i:])
	c.m.lines[i] = []rune{}
	_, h := c.v.screen.Size()
	if c.c.y >= h-1 {
		c.c.offset++
	} else {
		c.c.y++
	}
	c.c.x = 0
}

// Delete line
func (c *controller) deleteline() {
	i := c.c.y + c.c.offset
	if len(c.m.lines) > 1 {
		c.m.lines = append(c.m.lines[:i], c.m.lines[i+1:]...)
		if i >= len(c.m.lines) {
			if c.c.y > 0 {
				c.c.y--
			} else if c.c.offset > 0 {
				c.c.offset--
			}
		}
	} else {
		c.m.lines[0] = []rune{}
	}
	// Everytime we delete a line cursor x position is reseted
	c.c.x = 0
}

// Add a new line from the cursor current position.
// If the cursor is in the middle of a line, split that line.
func (c *controller) newlinesplit() {
	l := c.m.lines[c.c.y+c.c.offset]
	before := l[:c.c.x]
	after := l[c.c.x:]
	// Make sure we insert at least one empty char per new line
	if len(before) == 0 {
		before = []rune{}
	}
	if len(after) == 0 {
		after = []rune{}
	}
	i := c.c.y + c.c.offset + 1
	c.m.lines = append(c.m.lines, nil)
	copy(c.m.lines[i+1:], c.m.lines[i:])
	c.m.lines[i-1] = before
	c.m.lines[i] = after
	_, h := c.v.screen.Size()
	if c.c.y >= h-1 {
		c.c.offset++
	} else {
		c.c.y++
	}
	c.c.x = 0
}

// Handle rune
func (c *controller) handlerune(r rune, nfn func()) {
	switch c.m.mode {
	case NORMAL:
		nfn()
	case INSERT:
		c.put(r)
	case COMMAND:
		c.v.cl.put(r)
	}
}

// Handle event key
func (c *controller) handle(ev *tcell.EventKey) {
	// Reset lastkey everytime handle is called
	prevkey := c.lastkey
	if c.m.mode == NORMAL {
		c.lastkey = 0
	}

	switch ev.Key() {
	case tcell.KeyLeft:
		c.left()
	case tcell.KeyRight:
		c.right()
	case tcell.KeyUp:
		c.up()
	case tcell.KeyDown:
		c.down()
	case tcell.KeyCtrlQ:
		c.quit()
	case tcell.KeyCtrlC:
		c.m.mode = NORMAL
		c.v.cl.reset()
	case tcell.KeyEsc:
		switch c.m.mode {
		case NORMAL:
		case INSERT, COMMAND:
			if c.m.mode == INSERT {
				c.left()
			}
			c.m.mode = NORMAL
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		switch c.m.mode {
		case NORMAL:
			c.left()
		case INSERT:
			c.backspace()
		case COMMAND:
			c.v.cl.backspace()
		default:
			// Do nothing
		}
	case tcell.KeyEnter:
		switch c.m.mode {
		case NORMAL:
			c.down()
		case INSERT:
			c.newlinesplit()
		case COMMAND:
			c.exec()
		}
	}

	s := ev.Str()
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError || size != len(s) {
		return
	}
	switch r {
	case 'j':
		c.handlerune(r, c.down)
	case 'k':
		c.handlerune(r, c.up)
	case 'h':
		c.handlerune(r, c.left)
	case 'l':
		c.handlerune(r, c.right)
	case 'i':
		nfn := func() { c.m.mode = INSERT }
		c.handlerune(r, nfn)
	case 'o':
		nfn := func() {
			c.newline()
			c.m.mode = INSERT
		}
		c.handlerune(r, nfn)
	case 'd':
		nfn := func() {
			if prevkey == 'd' {
				c.deleteline()
			} else {
				c.lastkey = r
			}
		}
		c.handlerune(r, nfn)
	case 'e':
		c.handlerune(r, c.rightword)
	case 'b':
		c.handlerune(r, c.leftword)
	case 'x':
		c.handlerune(r, c.deletechar)
	case ':':
		nfn := func() {
			c.m.mode = COMMAND
			c.v.cl.put(r)
		}
		c.handlerune(r, nfn)
	default:
		c.handlerune(r, func() {})
	}
}

// Run editor main loop and poll key events.
func (c *controller) run() {
	for {
		ev := <-c.v.screen.EventQ()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			c.v.draw(c.m, &c.c)
		case *tcell.EventKey:
			c.handle(ev)
		default:
			c.v.draw(c.m, &c.c)
		}
		c.v.draw(c.m, &c.c)
	}
}

// Exec content from command buffer
func (c *controller) exec() {
	switch string(c.v.cl.buf) {
	case ":q":
		c.quit()
	case ":w":
		c.m.write()
	case ":wq":
		c.m.write()
		c.quit()
	}
	// Reset command buffer after command is executed
	c.v.cl.reset()
	// Put editor automatically in normal mode
	c.m.mode = NORMAL
}

// Quit editor
func (c *controller) quit() {
	c.v.screen.Fini()
	os.Exit(0)
}
