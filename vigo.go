package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide file name")
	}
    var m model
    var v view
    var c controller

    m.init()
    v.init()
	c.init(&m, &v)

	m.open(os.Args[1])
	c.run()
}
