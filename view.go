package main

import (
    "log"

	"github.com/gdamore/tcell/v3"
)

type view struct {
	screen  tcell.Screen
	style   tcell.Style
    cl      cmdl
}

func (v *view) init() {
    var err error
	v.screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	err = v.screen.Init()
	if err != nil {
		log.Fatal(err)
	}
	v.style = tcell.StyleDefault.Normal()
	_, h := v.screen.Size()
	v.cl.init(h)

}

// Draw editor content on screen
func (v *view) draw(m *model, c *cursor) {
	v.screen.Clear()
	w, h := v.screen.Size()
	for i := 0; i < h-1 && (i+c.offset) < len(m.lines); i++ {
		l := m.lines[i+c.offset]
		for j, c := range l {
			// Draw letters until we reach maximum width
			// Todo: implement a mode to do line wrap
			if j >= w {
				break
			}
			v.screen.SetContent(j, i, c, nil, v.style)
		}
	}
	for i, r := range v.cl.buf {
		if i >= w {
			break
		}
		v.screen.SetContent(i, h-1, r, nil, v.style)
	}
	if m.mode == NORMAL || m.mode == INSERT {
		v.screen.ShowCursor(c.x, c.y)
	} else if m.mode == COMMAND {
		v.screen.ShowCursor(v.cl.c.x, v.cl.c.y)
	}
	v.screen.Show()
}

