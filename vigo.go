package main

import (
	"bufio"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

const (
	NORMAL = iota
	INSERT
	COMMAND
)

type editor struct {
	fname  string
	c      cursor
	lines  []string
	screen tcell.Screen
	style  tcell.Style
	mode   int
}

type cursor struct {
	x      int
	y      int
	offset int
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
	e.mode = NORMAL
}

// Draw editor content on screen
func (e *editor) draw() {
	e.screen.Clear()
	//h, w := e.screen.Size()
	for y, l := range e.lines {
		for x, c := range l {
			e.screen.SetContent(x, y, c, nil, e.style)
		}
	}
	e.screen.ShowCursor(e.c.x, e.c.y)
	e.screen.Show()
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
			case tcell.KeyCtrlQ:
				e.screen.Fini()
				os.Exit(0)
			default:
			}
		default:
			e.draw()
		}
		e.draw()
	}
}
