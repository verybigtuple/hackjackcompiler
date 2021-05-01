package main

import (
	"fmt"
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

func (t *ParseTree) feedKeyword(val string) *KeywordToken {
	k, ok := t.current.(*KeywordToken)
	if ok && k.GetValue() == val {
		t.current, _ = t.tz.ReadToken()
	} else {
		t.errorf("Expected Keyword %s", val)
	}
	return k
}

func (t *ParseTree) feedIdent() *IdentifierToken {
	tk, ok := t.current.(*IdentifierToken)
	if ok {
		t.current, _ = t.tz.ReadToken()
	} else {
		t.errorf("Expected Identifier")
	}
	return tk
}

func (t *ParseTree) feedType() Token {
	switch tk := t.current.(type) {
	case *KeywordToken:
		val := tk.GetValue()
		if val == "int" || val == "char" || val == "bool" {
			t.current, _ = t.tz.ReadToken()
			return tk
		}
	case *IdentifierToken:
		t.current, _ = t.tz.ReadToken()
		return tk
	}
	t.errorf("Expected type")
	return nil
}

// varDec:'var' type varName (','varName)*';'
func (t *ParseTree) varDec() *VarDecNode {
	var err error
	t.current, err = t.tz.ReadToken()
	if err != nil {
		t.errorf("Unexpected item %v", err)
	}

	t.feedKeyword("var")
	vd := NewVarDecNode(t.feedType(), t.feedIdent())
	return vd
}
