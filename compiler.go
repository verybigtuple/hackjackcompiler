package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

type MemSegment string

const (
	ConstSegm   MemSegment = "constant"
	LocalSegm              = "local"
	ArgSegm                = "argument"
	ThisSegm               = "this"
	TempSegm               = "temp"
	StaticSegm             = "static"
	PointerSegm            = "pointer"
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

var binaryOps = map[string]string{
	"+": "add",
	"-": "sub",
	"&": "and",
	"|": "or",
	"=": "eq",
	">": "gt",
	"<": "lt",
}

// Operations that should use OS functions
var sysBinaryOps = map[string]string{
	"*": "Math.multiply",
	"/": "Math.divide",
}

var unaryOps = map[string]string{
	"~": "not",
	"-": "neg",
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

func (c *Compiler) errorf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}

func (c *Compiler) error(msg string) {
	c.errorf("%s", msg)
}

func (c *Compiler) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		*errp = e.(error)
	}
}

func (c *Compiler) Run(root Node) (err error) {
	defer c.recover(&err)
	root.Compile(c)
	return
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
	c.sb.WriteString("return\n")
}

func (c *Compiler) BinaryOp(symbol string) {
	if cmd, ok := binaryOps[symbol]; ok {
		c.sb.WriteString(cmd + "\n")
	} else if sf, ok := sysBinaryOps[symbol]; ok {
		c.Call(sf, 2)
	} else {
		c.error("Undefined binary op")
	}
}

func (c *Compiler) UnaryOp(symbol string) {
	if cmd, ok := unaryOps[symbol]; ok {
		c.sb.WriteString(cmd + "\n")
	} else {
		c.error("Undefined unary op")
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
func (c *Compiler) OpenWhile() (begin, end string) {
	fn := c.SymbolTblList.Name()
	begin = fn + "$WHILE_BEGIN_" + strconv.Itoa(c.whileCount)
	end = fn + "$WHILE_END_" + strconv.Itoa(c.whileCount)
	c.whileCount++
	return
}

func (c *Compiler) OpenIf() (els, end string) {
	fn := c.SymbolTblList.Name()
	els = fn + "$ELSE_" + strconv.Itoa(c.ifCount)
	end = fn + "$IF_END_" + strconv.Itoa(c.ifCount)
	c.ifCount++
	return
}
