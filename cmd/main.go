package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"

	"github.com/Dyleme/ch-graph/pkg/printer"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

const hello = `
package main

import "fmt"

const message = "Hello, World!"

func sendToCh(ch chan string) {
	ch <- message
}

func main() {
	ch := make(chan string)

	go sendToCh(ch)

	ch2 := make(chan int)

	select {
	case str := <-ch:
		fmt.Println(str)
	case <-ch2:
		fmt.Println("default")
	}
}
`

func main() {
	// Replace interface{} with any for this test.
	// ssa.SetNormalizeAnyForTesting(true)
	// defer ssa.SetNormalizeAnyForTesting(false)
	// Parse the source files.
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "hello.go", hello, parser.ParseComments)
	if err != nil {
		fmt.Print(err) // parse error
		return
	}
	files := []*ast.File{f}

	// Create the type-checker's package.
	pkg := types.NewPackage("hello", "")

	// Type-check the package, load dependencies.
	// Create and build the SSA program.
	hello, _, err := ssautil.BuildPackage(
		&types.Config{Importer: importer.Default()}, fset, pkg, files, ssa.SanityCheckFunctions)
	if err != nil {
		fmt.Print(err) // type error in some package
		return
	}

	// Print out the package.
	hello.WriteTo(os.Stdout)

	// Print out the package-level functions.
	// hello.Func("init").WriteTo(os.Stdout)
	// hello.Func("main").WriteTo(os.Stdout)

	for i, member := range hello.Members {
		printer.Println(i, member, member.Token())
		if member.Token().String() == "func" {
			f := hello.Func(member.Name())
			params := f.Params
			printer.Println("params:")
			printer.IncreaseLevel()
			for i, p := range params {
				fmt.Printf("\tp%v %v\n", i, p)
				printInstructions(*p.Referrers())
			}
			printer.DecreaseLevel()
			for _, b := range f.Blocks {
				printInstructions(b.Instrs)
			}

		}
		fmt.Println() // fmt.Println(member.Type())
	}
}

func printInstructions(instrs []ssa.Instruction) {
	printer.IncreaseLevel()
	defer printer.DecreaseLevel()
	for i, instr := range instrs {
		printer.Println(i, instr)
		operands := make([]*ssa.Value, 0)
		operands = instr.Operands(operands)
		printer.Println("operands:")
		printer.IncreaseLevel()
		for i, op := range operands {
			o := *op
			if o == nil {
				printer.Printf("\top%v %v type:<nil>\n", i, o)
			} else {
				printer.Printf("\top%v %v name: %v type:%v\n", i, o, o.Name(), o.Type())
			}
		}
		printer.DecreaseLevel()
	}
}
