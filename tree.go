package main

import (
	"fmt"
	"io"
	"runtime"
)

type ParseTree struct {
	tz      *Tokenizer
	current Token
}

func NewPasreTree(tz *Tokenizer) *ParseTree {
	return &ParseTree{tz: tz, current: nil}
}

func (t *ParseTree) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		*errp = e.(error)
	}
}

func (t *ParseTree) errorf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}

func (t *ParseTree) Parse() (root Node, err error) {
	defer t.recover(&err)
	root = t.varDec()
	return
}

func (t *ParseTree) feedToken(tt TokenType, val string) Token {
	tk := t.current
	if tk.Type() != tt {
		t.errorf("Unexpexted token type. Expected %v; Got: %v", tt, tk.Type())
	}
	if val != "" && tk.GetValue() != val {
		t.errorf("Unexpected token value. Expected %s; Got: %s", val, tk.GetValue())
	}
	var err error
	t.current, err = t.tz.ReadToken()
	if err != nil && err != io.EOF {
		t.errorf("Unexpected next token")
	}
	return tk
}

// type:'int'|'char'|'boolean'|className
func (t *ParseTree) feedType() Token {
	tk := t.current
	switch tk.Type() {
	case TokenKeyword:
		val := tk.GetValue()
		if val != "int" && val != "char" && val != "bool" {
			t.errorf("Unexpected type")
		}
	case TokenIdentifier:
		break
	default:
		t.errorf("Unexpected Token for Var type")
	}

	var err error
	t.current, err = t.tz.ReadToken()
	if err != nil {
		t.errorf("Unexpected next token")
	}

	return tk
}

// varDec:'var' type varName (','varName)*';'
func (t *ParseTree) varDec() *VarDecNode {
	var err error
	t.current, err = t.tz.ReadToken()
	if err != nil {
		t.errorf("Unexpected item %v", err)
	}

	t.feedToken(TokenKeyword, "var")
	vd := NewVarDecNode(t.feedType(), t.feedToken(TokenIdentifier, ""))
	t.feedToken(TokenSymbol, ";")
	return vd
}
