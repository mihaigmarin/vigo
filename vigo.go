package main

import (
	"bufio"
	"log"
	"os"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

const (
	Normal = iota
	Insert
	Command
)

type cursor struct {
	x      int
	y      int
	offset int
}

// Cursor struct. Stores cursor position.
func (c *cursor) init() {
	c.x = 0
	c.y = 0
	c.offset = 0
}

// Command line struct. Commands are stored here.
type cmdl struct {
	buf string
}

// Initialize command line
func (cl *cmdl) init() {
	cl.buf = ""
}

// Put rune to command buffer.
func (cl *cmdl) put(r rune) {
	cl.buf += string(r)
}

// Reset command buffer.
func (cl *cmdl) reset() {
	cl.buf = ""
}

type editor struct {
	fname   string
	c       cursor
	lines   [][]rune
	screen  tcell.Screen
	style   tcell.Style
	mode    int
	s       *bufio.Scanner
	w       *bufio.Writer
	cl      cmdl
	lastkey rune
}

// Init editor
func (e *editor) init() {
	var err error
	e.c.init()
	e.cl.init()
	e.fname = ""
	e.lines = make([][]rune, 0)
	e.screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	err = e.screen.Init()
	if err != nil {
		log.Fatal(err)
	}
	e.style = tcell.StyleDefault.Normal()
	e.mode = Normal
	e.s = nil
	e.w = nil
}

// Open file path and read the content inside editor
func (e *editor) open(fname string) {
	e.fname = fname
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	e.s = bufio.NewScanner(f)
	for e.s.Scan() {
		l := e.s.Text()
		e.lines = append(e.lines, []rune(l))
	}
	// If the file doesn't have any lines,
	// add empty space at the start of line
	if len(e.lines) == 0 {
		e.lines = append(e.lines, []rune{' '})
	}
}

// Write lines to the current file opened by the editor.
func (e *editor) write() {
	f, err := os.OpenFile(e.fname, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	e.w = bufio.NewWriter(f)
	for _, l := range e.lines {
		for _, r := range l {
			_, err := e.w.WriteRune(r)
			if err != nil {
				log.Fatal(err)
			}
		}
		_, err = e.w.WriteRune('\n')
		if err != nil {
			log.Fatal(err)
		}
	}
	err = e.w.Flush()
	if err != nil {
		log.Fatal(err)
	}
}

// Draw editor content on screen
func (e *editor) draw() {
	e.screen.Clear()
	w, h := e.screen.Size()
	for i := 0; i < h-1 && (i+e.c.offset) < len(e.lines); i++ {
		l := e.lines[i+e.c.offset]
		for j, c := range l {
			// Draw letters until we reach maximum width
			// Todo: implement a mode to do line wrap
			if j >= w {
				break
			}
			e.screen.SetContent(j, i, c, nil, e.style)
		}
	}
	for i, c := range e.cl.buf {
		if i >= w {
			break
		}
		e.screen.SetContent(i, h-1, c, nil, e.style)
	}
	e.screen.ShowCursor(e.c.x, e.c.y)
	e.screen.Show()
}

// Move editor cursor up.
func (e *editor) up() {
	if e.c.y > 0 {
		e.c.y--
	} else if e.c.offset > 0 {
		e.c.offset--
	}
	if e.c.x > len(e.lines[e.c.y+e.c.offset])-1 {
		e.c.x = len(e.lines[e.c.y+e.c.offset]) - 1
	}
}

// Move editor cursor down.
func (e *editor) down() {
	_, h := e.screen.Size()
	if e.c.y+e.c.offset <= len(e.lines)-2 {
		if e.c.y < h-2 {
			e.c.y++
		} else {
			e.c.offset++
		}
		if e.c.x > len(e.lines[e.c.y+e.c.offset])-1 {
			e.c.x = len(e.lines[e.c.y+e.c.offset]) - 1
		}
	}
}

// Move editor cursor left.
func (e *editor) left() {
	if e.c.x > 0 {
		e.c.x--
	}
}

// Move editor cursor right.
func (e *editor) right() {
	if e.c.x < len(e.lines[e.c.y+e.c.offset])-1 {
		e.c.x++
	}
}

// put rune to screen
func (e *editor) put(r rune) {
	// put only printable character, avoid control characters such
	// Enter, Backspace, etc. Those are handled separatly in the
	// function 'handleKey'
	if !unicode.IsControl(r) {
		pl := &e.lines[e.c.y+e.c.offset]
		*pl = append((*pl)[:e.c.x], r)
		*pl = append(*pl, (*pl)[e.c.x+1:]...)
		e.c.x++
	}
}

// Delete rune from screen
func (e *editor) delete() {
	if e.c.x > 0 {
		pl := &e.lines[e.c.y+e.c.offset]
		*pl = append((*pl)[:e.c.x-1], (*pl)[e.c.x:]...)
		e.c.x--
	}
}

// Add a new line
func (e *editor) newline() {
	i := e.c.y + e.c.offset + 1
	e.lines = append(e.lines, nil)
	copy(e.lines[i+1:], e.lines[i:])
	e.lines[i] = []rune{}
	_, h := e.screen.Size()
	if e.c.y >= h-1 {
		e.c.offset++
	} else {
		e.c.y++
	}
	e.c.x = 0
}

// Delete line
func (e *editor) deleteline() {
	i := e.c.y + e.c.offset
	e.lines = append(e.lines[:i], e.lines[i+1:]...)
	if i == len(e.lines) {
		e.c.y--
	}
	// Everytime we delete a line cursor x position is reseted
	e.c.x = 0
}

// Add a new line from the cursor current position.
// If the cursor is in the middle of a line, split that line.
func (e *editor) newlinesplit() {
	l := e.lines[e.c.y+e.c.offset]
	before := l[:e.c.x]
	after := l[e.c.x:]
	// Make sure we insert at least one empty char per new line
	if len(before) == 0 {
		before = []rune{}
	}
	if len(after) == 0 {
		after = []rune{}
	}
	i := e.c.y + e.c.offset + 1
	e.lines = append(e.lines, nil)
	copy(e.lines[i+1:], e.lines[i:])
	e.lines[i-1] = before
	e.lines[i] = after
	_, h := e.screen.Size()
	if e.c.y >= h-1 {
		e.c.offset++
	} else {
		e.c.y++
	}
	e.c.x = 0
}

// Handle event key
func (e *editor) handle(ev *tcell.EventKey) {
	// Reset lastkey everytime handle is called
	prevkey := e.lastkey
	if e.mode == Normal {
		e.lastkey = 0
	}

	switch ev.Key() {
	case tcell.KeyLeft:
		e.left()
	case tcell.KeyRight:
		e.right()
	case tcell.KeyUp:
		e.up()
	case tcell.KeyDown:
		e.down()
	case tcell.KeyCtrlQ:
		e.quit()
	case tcell.KeyEsc:
		switch e.mode {
		case Normal:
		case Insert, Command:
			e.mode = Normal
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		switch e.mode {
		case Normal:
			e.left()
		case Insert:
			e.delete()
		default:
			// Do nothing
		}
	case tcell.KeyEnter:
		switch e.mode {
		case Normal:
			e.down()
		case Insert:
			e.newlinesplit()
		case Command:
			e.exec()
		}
	}

	switch ev.Rune() {
	case 'j':
		switch e.mode {
		case Normal:
			e.down()
		case Insert:
			e.put(ev.Rune())
		case Command:
		}
	case 'k':
		switch e.mode {
		case Normal:
			e.up()
		case Insert:
			e.put(ev.Rune())
		case Command:
		}
	case 'h':
		switch e.mode {
		case Normal:
			e.left()
		case Insert:
			e.put(ev.Rune())
		case Command:
		}
	case 'l':
		switch e.mode {
		case Normal:
			e.right()
		case Insert:
			e.put(ev.Rune())
		case Command:
		}
	case 'i':
		switch e.mode {
		case Normal:
			e.mode = Insert
		case Insert:
			e.put(ev.Rune())
		case Command:
		}
	case 'o':
		switch e.mode {
		case Normal:
			e.newline()
			e.mode = Insert
		case Insert:
			e.put(ev.Rune())
		case Command:
		}
	case 'd':
		switch e.mode {
		case Normal:
			if prevkey == 'd' {
				e.deleteline()
			} else {
				e.lastkey = ev.Rune()
			}
		case Insert:
			e.put(ev.Rune())
		case Command:
		}
	case ':':
		switch e.mode {
		case Normal:
			e.mode = Command
			e.cl.put(ev.Rune())
		case Insert:
		case Command:
			e.cl.put(ev.Rune())
		}
	default:
		switch e.mode {
		case Normal:
		case Insert:
			e.put(ev.Rune())
		case Command:
			e.cl.put(ev.Rune())
		}
	}
}

// Run editor main loop and poll key events.
func (e *editor) run() {
	for {
		ev := e.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			e.draw()
		case *tcell.EventKey:
			e.handle(ev)
		default:
			e.draw()
		}
		e.draw()
	}
}

// Exec content from command buffer
func (e *editor) exec() {
	switch e.cl.buf {
	case ":q":
		e.quit()
	case ":w":
		e.write()
	case ":wq":
		e.write()
		e.quit()
	}
	// Reset command buffer after command is executed
	e.cl.reset()
	// Put editor automatically in normal mode
	e.mode = Normal
}

// Quit editor
func (e *editor) quit() {
	e.screen.Fini()
	os.Exit(0)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide file name")
	}
	var e editor
	e.init()
	e.open(os.Args[1])
	e.run()
}
