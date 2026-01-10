package main

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
