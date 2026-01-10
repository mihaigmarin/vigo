package main

import (
    "bufio"
    "os"
    "log"
)

const (
    NORMAL = iota
    INSERT
    COMMAND
)

type model struct {
	fname   string
	s       *bufio.Scanner
	w       *bufio.Writer
    lines   [][]rune
    mode    int
}

func (m *model) init() {
    m.fname = ""
    m.s = nil
    m.w = nil
    m.lines = make([][]rune, 0)
    m.mode = NORMAL
}

// Open file path and read the content inside editor
func (m *model) open(fname string) {
	m.fname = fname
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	m.s = bufio.NewScanner(f)
	for m.s.Scan() {
		l := m.s.Text()
		m.lines = append(m.lines, []rune(l))
	}
	// If the file doesn't have any lines,
	// add empty space at the start of line
	if len(m.lines) == 0 {
		m.lines = append(m.lines, []rune{' '})
	}
}

// Write lines to the current file opened by the editor.
func (m *model) write() {
	f, err := os.OpenFile(m.fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	m.w = bufio.NewWriter(f)
	for _, l := range m.lines {
		for _, r := range l {
			_, err := m.w.WriteRune(r)
			if err != nil {
				log.Fatal(err)
			}
		}
		_, err = m.w.WriteRune('\n')
		if err != nil {
			log.Fatal(err)
		}
	}
	err = m.w.Flush()
	if err != nil {
		log.Fatal(err)
	}
}

