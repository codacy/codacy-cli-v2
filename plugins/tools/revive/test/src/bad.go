package testdata

import (
	"fmt"
)

// This function has too many arguments, a magic number, and a bare return
func BadFunction(a int, b int, c int, d int, e int) int {
	var foo int = 42 // magic number
	if foo == 42 {
		fmt.Println("foo is 42")
		return 56
	}
	return foo
}

// This function has a confusing name and an unused parameter
func xYz(_ int, unused int) {
	fmt.Println("confusing name and unused param")
}

// This function has a long line
func LongLine() {
	fmt.Println("This is a very very very very very very very very very very very very very very very very very very very long line")
}

// This function has a naked return
func NakedReturn() (x int) {
	return
}
