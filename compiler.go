package main

import (
	"strconv"
	"strings"
)

type MemSegment string

const (
	ConstSegm  MemSegment = "constant"
	LocalSegm             = "local"
	ArgSegm               = "argument"
	ThisSegm              = "this"
	TempSegm              = "temp"
	StaticSegm            = "static"
)

var vKinds = map[VarKind]MemSegment{
	Field:  ThisSegm,
	Arg:    ArgSegm,
	Local:  LocalSegm,
	Static: StaticSegm,
}

func GetSegment(vk VarKind) MemSegment {
	return vKinds[vk]
}

var ops = map[string]string{
	"+": "add",
	"-": "sub",
	"&": "and",
	"|": "or",
	"=": "eq",
	">": "gt",
	"<": "lt",
	"~": "not",
}

// Operations that should use OS functions
var sysOps = map[string]string{
	"*": "Math.multiply",
	"/": "Math.divide",
}

type Compiler struct {
	sb            *strings.Builder
	whileCount    int
	ifCount       int
	SymbolTblList *SymbolTableList
}

func NewCompiler() *Compiler {
	stList := NewSymbolTableList()
	sb := &strings.Builder{}
	return &Compiler{sb: sb, SymbolTblList: stList}
}

func (c *Compiler) String() string {
	return c.sb.String()
}

func (c *Compiler) Push(segm MemSegment, offset string) {
	c.sb.WriteString("push " + string(segm) + " " + offset + "\n")
}

func (c *Compiler) Pop(segm MemSegment, offset string) {
	c.sb.WriteString("pop " + string(segm) + " " + offset + "\n")
}

func (c *Compiler) Function(name string, localVarCount int) {
	c.sb.WriteString("function " + name + " " + strconv.Itoa(localVarCount) + "\n")
}

func (c *Compiler) Call(name string, argsCount int) {
	c.sb.WriteString("call " + name + " " + strconv.Itoa(argsCount) + "\n")
}

func (c *Compiler) Return() {
	c.sb.WriteString("return \n")
}

func (c *Compiler) Op(symbol string) {
	if cmd, ok := ops[symbol]; ok {
		c.sb.WriteString(cmd + "\n")
	} else if sf, ok := sysOps[symbol]; ok {
		c.Call(sf, 2)
	}
}

func (c *Compiler) Label(name string) {
	c.sb.WriteString("label " + name + "\n")
}

func (c *Compiler) Goto(label string) {
	c.sb.WriteString("goto " + label + "\n")
}

func (c *Compiler) IfGoto(label string) {
	c.sb.WriteString("if-goto " + label + "\n")
}

// OpenWhile returns 2 label names for beginWhile and endWhile
func (c *Compiler) OpenWhile() (begin string, end string) {
	fn := strings.ToUpper(c.SymbolTblList.Name())
	begin = fn + "_WHILE_BEGIN_" + strconv.Itoa(c.whileCount)
	end = fn + "_WHILE_END_" + strconv.Itoa(c.whileCount)
	return
}

func (c *Compiler) CloseWhile() {
	c.whileCount++
}
