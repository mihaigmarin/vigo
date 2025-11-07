package main

import (
	"bufio"
	"log"
	"os"

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

// Draw editor content on screen
func (e *editor) draw() {
	e.screen.Clear()
	w, h := e.screen.Size()
	for i := 0; i < h && (i + e.c.offset) < len(e.lines); i++ {
		l := e.lines[i + e.c.offset]
		for j, c := range l {
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
		if e.c.x > len(e.lines[e.c.y + e.c.offset]) - 1 {
			e.c.x = len(e.lines[e.c.y + e.c.offset]) - 1
		}
	}
}

// Move editor cursor down.
func (e *editor) down() {
	_, h := e.screen.Size()
	if e.c.y + e.c.offset <= len(e.lines) - 2 {
		if e.c.y < h - 1 {
			e.c.y++
		} else {
			e.c.offset++
		}
		if e.c.x > len(e.lines[e.c.y + e.c.offset]) {
			e.c.x = len(e.lines[e.c.y + e.c.offset]) - 1
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
	if e.c.x < len(e.lines[e.c.y + e.c.offset]) - 1 {
		e.c.x++
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide file name")
	}
	fname := os.Args[1]
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var e editor
	e.init()
	s := bufio.NewScanner(f)
	for s.Scan() {
		l := s.Text()
		e.lines = append(e.lines, l)
	}
	for {
		ev := e.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			e.draw()
		case *tcell.EventKey:
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
				e.screen.Fini()
				os.Exit(0)
			default:
				// Do nothing
			}
		default:
			e.draw()
		}
		e.draw()
	}
}
