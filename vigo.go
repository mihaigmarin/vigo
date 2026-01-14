package main

import (
	"bufio"
	"log"
	"os"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
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
	_, h := e.screen.Size()
	e.cl.init(h)
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
		e.lines = append(e.lines, []rune{})
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
	// Note: Add '~' on each remaining line on the screen that
	// is not contained inside "e.lines".
	for i := len(e.lines); i < h-1; i++ {
		e.screen.SetContent(0, i, rune('~'), nil, e.style)
	}
	for i, r := range e.cl.buf {
		if i >= w {
			break
		}
		e.screen.SetContent(i, h-1, r, nil, e.style)
	}
	if e.mode == Normal || e.mode == Insert {
		e.screen.ShowCursor(e.c.x, e.c.y)
	} else if e.mode == Command {
		e.screen.ShowCursor(e.cl.c.x, e.cl.c.y)
	}
	e.screen.Show()
}

// Returns the line limit at the "y" position.
func (e *editor) linelimit(y int) int {
	limit := len(e.lines[y]) - 1
	if e.mode == Insert {
		limit = len(e.lines[y])
	}
	if limit < 0 {
		limit = 0
	}
	return limit
}

// Move editor cursor up.
func (e *editor) up() {
	if e.c.y > 0 {
		e.c.y--
	} else if e.c.offset > 0 {
		e.c.offset--
	}
	limit := e.linelimit(e.c.y + e.c.offset)
	if e.c.x > limit {
		e.c.x = limit
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
		limit := e.linelimit(e.c.y + e.c.offset)
		if e.c.x > limit {
			e.c.x = limit
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
	limit := e.linelimit(e.c.y + e.c.offset)
	if e.c.x < limit {
		e.c.x++
	}
}

// Move cursor left until the end of word. Triggered with "b" in normal mode
func (e *editor) leftword() {
	i := e.c.y + e.c.offset

	e.left()
	for e.c.x >= 0 && unicode.IsSpace(e.lines[i][e.c.x]) {
		e.left()
	}
	for e.c.x-1 >= 0 {
		if unicode.IsSpace(e.lines[i][e.c.x-1]) {
			break
		}
		e.left()
	}
}

// Move cursor right until the end of word. Triggered with "e" in normal mode
func (e *editor) rightword() {
	i := e.c.y + e.c.offset

	e.right()
	for e.c.x < len(e.lines[i]) && unicode.IsSpace(e.lines[i][e.c.x]) {
		e.right()
	}
	for e.c.x+1 < len(e.lines[i]) {
		if unicode.IsSpace(e.lines[i][e.c.x+1]) {
			break
		}
		e.right()
	}
}

// Move cursor to the end of line.
func (e *editor) endofline() {
	limit := e.linelimit(e.c.y + e.c.offset)
	e.c.x = limit
}

// Move cursor to the start of line.
func (e *editor) startofline() {
	e.c.x = 0
}

// put rune to screen
func (e *editor) put(r rune) {
	// put only printable character, avoid control characters such
	// Enter, Backspace, etc. Those are handled separatly in the
	// function 'handleKey'
	if !unicode.IsControl(r) {
		pl := &e.lines[e.c.y+e.c.offset]
		*pl = append((*pl)[:e.c.x], append([]rune{r}, (*pl)[e.c.x:]...)...)
		e.c.x++
	}
}

// Handle backspace press in insert mode. When triggered this function
// deletes the rune after the cursor position in the editor buffer.
func (e *editor) backspace() {
	if e.c.x > 0 {
		pl := &e.lines[e.c.y+e.c.offset]
		*pl = append((*pl)[:e.c.x-1], (*pl)[e.c.x:]...)
		e.c.x--
	}
}

func (e *editor) deletechar() {
	i := e.c.y + e.c.offset
	pl := &e.lines[i]

	if len(*pl) == 0 {
		return
	}

	if len(*pl) == 1 {
		*pl = []rune{}
		e.c.x = 0
		return
	}

	if e.c.x < len(*pl) {
		*pl = append((*pl)[:e.c.x], (*pl)[e.c.x+1:]...)
		if e.c.x >= len(*pl) {
			e.c.x = len(*pl) - 1
		}
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
	before := make([]rune, e.c.x)
	after := make([]rune, len(l)-e.c.x)
	copy(before, l[:e.c.x])
	copy(after, l[e.c.x:])
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

// Handle rune
func (e *editor) handlerune(r rune, nfn func()) {
	switch e.mode {
	case Normal:
		nfn()
	case Insert:
		e.put(r)
	case Command:
		e.cl.put(r)
	}
}

// Handle event key
func (e *editor) handle(ev *tcell.EventKey) {
	// Reset lastkey everytime handle is called
	prevkey := e.lastkey
	if e.mode == Normal {
		e.lastkey = 0
	}

	if ev.Key() != tcell.KeyRune {
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
		case tcell.KeyCtrlC:
			e.mode = Normal
			e.cl.reset()
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
				e.backspace()
			case Command:
				e.cl.backspace()
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
		// Note: Return if we handled a special key that is not a rune.
		// In this way, we will not continue with the rune logic.
		return
	}

	s := ev.Str()
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError || size != len(s) {
		return
	}
	switch r {
	case 'j':
		e.handlerune(r, e.down)
	case 'k':
		e.handlerune(r, e.up)
	case 'h':
		e.handlerune(r, e.left)
	case 'l':
		e.handlerune(r, e.right)
	case 'i':
		nfn := func() { e.mode = Insert }
		e.handlerune(r, nfn)
	case 'o':
		nfn := func() {
			e.newline()
			e.mode = Insert
		}
		e.handlerune(r, nfn)
	case 'd':
		nfn := func() {
			if prevkey == 'd' {
				e.deleteline()
			} else {
				e.lastkey = r
			}
		}
		e.handlerune(r, nfn)
	case 'e':
		e.handlerune(r, e.rightword)
	case 'b':
		e.handlerune(r, e.leftword)
	case 'x':
		e.handlerune(r, e.deletechar)
	case '$':
		e.handlerune(r, e.endofline)
	case '0':
		e.handlerune(r, e.startofline)
	case ':':
		nfn := func() {
			e.mode = Command
			e.cl.put(r)
		}
		e.handlerune(r, nfn)
	default:
		e.handlerune(r, func() {})
	}
}

// Run editor main loop and poll key events.
func (e *editor) run() {
	for {
		ev := <-e.screen.EventQ()
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
	switch string(e.cl.buf) {
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
