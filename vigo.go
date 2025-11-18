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

type editor struct {
	fname  string
	c      cursor
	lines  []string
	screen tcell.Screen
	style  tcell.Style
	mode   int
	s      *bufio.Scanner
	f      *os.File
}

// Init editor
func (e *editor) init() {
	var err error
	e.fname = ""
	e.c.x = 0
	e.c.y = 0
	e.c.offset = 0
	e.lines = make([]string, 0)
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
}

// Open file path and read the content inside editor
func (e *editor) open(fname string) {
	var err error
	e.f, err = os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer e.f.Close()
	e.s = bufio.NewScanner(e.f)
	for e.s.Scan() {
		l := e.s.Text()
		e.lines = append(e.lines, l)
	}
}

// Draw editor content on screen
func (e *editor) draw() {
	e.screen.Clear()
	w, h := e.screen.Size()
	for i := 0; i < h && (i+e.c.offset) < len(e.lines); i++ {
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
	e.screen.ShowCursor(e.c.x, e.c.y)
	e.screen.Show()
}

// Move editor cursor up.
func (e *editor) up() {
	if e.c.y > 0 {
		e.c.y--
		if e.c.x > len(e.lines[e.c.y+e.c.offset])-1 {
			e.c.x = len(e.lines[e.c.y+e.c.offset]) - 1
		}
	}
}

// Move editor cursor down.
func (e *editor) down() {
	_, h := e.screen.Size()
	if e.c.y+e.c.offset <= len(e.lines)-2 {
		if e.c.y < h-1 {
			e.c.y++
		} else {
			e.c.offset++
		}
		if e.c.x > len(e.lines[e.c.y+e.c.offset]) {
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

// Write rune to screen
func (e *editor) write(r rune) {
	// Write only printable character, avoid control characters such
	// Enter, Backspace, etc. Those are handled separatly in the
	// function 'handleKey'
	if !unicode.IsControl(r) {
		pl := &e.lines[e.c.y+e.c.offset];
		*pl = (*pl)[:e.c.x] + string(r) + (*pl)[e.c.x:]
		e.c.x++
	}
}

// Delete rune from screen
func (e *editor) delete() {
	if e.c.x > 0 {
		pl := &e.lines[e.c.y+e.c.offset];
		*pl = (*pl)[:e.c.x-1] + (*pl)[e.c.x:]
		e.c.x--
	}
}

// Handle rune from stdin
func (e *editor) handleRune(ev *tcell.EventKey) {
	switch e.mode {
	case Normal:
		switch ev.Rune() {
		case 'j':
			e.down()
		case 'k':
			e.up()
		case 'h':
			e.left()
		case 'l':
			e.right()
		case 'i':
			e.mode = Insert
		}
	case Insert:
		e.write(ev.Rune())
	case Command:
	default:
		// Do nothing
	}
}

// Handle key from stdin
func (e *editor) handleKey(ev *tcell.EventKey) {
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
		e.mode = Normal
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		switch e.mode {
		case Normal:
			e.left()
		case Insert:
			e.delete()
		case Command:
		default:
			// Do nothing
		}
	default:
		// Do nothing
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
			e.handleKey(ev)
			e.handleRune(ev)
		default:
			e.draw()
		}
		e.draw()
	}
}

// Move editor cursor right.
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
