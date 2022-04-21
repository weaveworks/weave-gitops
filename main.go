package main

import (
	"fmt"
)

func main() {
}

func test() string {
	a := 1
	w := 21
	x := 6
	y := 28
	z := 496
	q := 8128

	return fmt.Sprintf("%s:%d", "test", a*w*x*y*z*q)
}
