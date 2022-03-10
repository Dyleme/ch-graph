package printer

import "fmt"

var level int

func IncreaseLevel() {
	level++
}

func DecreaseLevel() {
	level--
}

func Println(a ...interface{}) {
	printLevel()
	fmt.Println(a...)
}

func Printf(format string, a ...interface{}) {
	printLevel()
	fmt.Printf(format, a...)
}

func printLevel() {
	for i := 0; i < level; i++ {
		fmt.Print("\t")
	}
}
