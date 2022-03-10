package functions

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Dyleme/ch-graph/pkg/printer"
	"golang.org/x/tools/go/ssa"
)

type Instruct interface {
	ssa.Instruction
	Type() string
}

type Program struct {
	FunctionNumber map[string]int
	Functions      []Function
	Chans          []ChanEdge
}

type Function struct {
	Blocks     map[int][]int
	ChName     map[string]string
	Name       string
	Operations Instruct
	Chans      []ChanEdge
}

type ChanEdge struct {
	Name string
}

func CreateFunction(f *ssa.Function) *Function {
	fmt.Println(f.Name())
	chans := make([]ChanEdge, 0)
	for _, p := range f.Params {
		if strings.Contains(p.Type().String(), "chan ") {
			// fmt.Println(i, p.Type())
			chans = append(chans, ChanEdge{p.Name()})
		}
	}

	fn := &Function{
		Name:   f.Name(),
		Chans:  chans,
		Blocks: make(map[int][]int),
	}

	fn.parseBlock(f, 0)

	fmt.Println(fn.Blocks)
	fmt.Println()
	return fn
}

func CreateFunctions(pack *ssa.Package) {
	functions := make([]*Function, 0)
	for _, m := range pack.Members {
		if m.Token().String() == "func" {
			f := pack.Func(m.Name())
			functions = append(functions, CreateFunction(f))
		}
	}
}

var (
	ifReg       = regexp.MustCompile(`^if [\w\d: ]+ goto (?P<then>\d+) else (?P<else>\d+)$`)
	jumpReg     = regexp.MustCompile(`^jump (?P<jump>\d+)&`)
	makeChanReg = regexp.MustCompile(`^make chan&`)
)

func (fn *Function) parseBlock(f *ssa.Function, number int) {
	if _, exist := fn.Blocks[number]; exist {
		return
	}
	fn.Blocks[number] = make([]int, 0)
	fmt.Println(number, "block")
	for i, instr := range f.Blocks[number].Instrs {
		printer.Println(i, instr)

		if matches := ifReg.FindStringSubmatch(instr.String()); matches != nil {
			thenIndex := ifReg.SubexpIndex("then")
			thenBlock, _ := strconv.Atoi(matches[thenIndex])
			elseIndex := ifReg.SubexpIndex("else")
			elseBlock, _ := strconv.Atoi(matches[elseIndex])
			fn.Blocks[number] = append(fn.Blocks[number], thenBlock)
			fn.Blocks[number] = append(fn.Blocks[number], elseBlock)
			fn.parseBlock(f, thenBlock)
			fn.parseBlock(f, elseBlock)
		}
		fmt.Println(instr.String())
		if matches := jumpReg.FindStringSubmatch(instr.String()); matches != nil {
			fmt.Println(`jumpBlock`)
			jumpIndex := ifReg.SubexpIndex("jump")
			jumpBlock, _ := strconv.Atoi(matches[jumpIndex])
			fn.Blocks[number] = append(fn.Blocks[number], jumpBlock)
			fn.parseBlock(f, jumpBlock)
		}
		// printer.Println("operands:")
		// operands := make([]*ssa.Value, 0)
		// operands = instr.Operands(operands)
		// printer.IncreaseLevel()
		// for i, op := range operands {
		// 	o := *op
		// 	if o == nil {
		// 		printer.Printf("\top%v %v type:<nil>\n", i, o)
		// 	} else {
		// 		printer.Printf("\top%v %v name: %v type:%v\n", i, o, o.Name(), o.Type())
		// 	}
		// }
		// printer.DecreaseLevel()
	}
}
