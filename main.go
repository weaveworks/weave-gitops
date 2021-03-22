package main

import (
	"fmt"
)

func main() {
}

func test() string {
	x := 6
	y := 28
	z := 496
	q := 8128
	return fmt.Sprintf("%s:%d", "test", x*y*z*q)
}
