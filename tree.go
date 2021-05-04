package main

import (
	"errors"
	"fmt"
	"io"
	"runtime"
)

type ParseTree struct {
	tz      *Tokenizer
	current Token
	peeked  Token
	root    func(*ParseTree) Node
}

func NewPasreTree(tz *Tokenizer) *ParseTree {
	pt := ParseTree{tz: tz}
	pt.root = func(pt *ParseTree) Node { return pt.letStatement() }
	return &pt
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
	root = t.root(t)
	return
}

func (t *ParseTree) next() Token {
	if t.peeked != nil {
		t.current, t.peeked = t.peeked, nil
		return t.current
	}

	var err error
	t.current, err = t.tz.ReadToken()
	if err != nil && !errors.Is(err, io.EOF) {
		t.errorf("Unexpected token: %v", err)
	}
	if err != nil && errors.Is(err, io.EOF) {
		t.errorf("Unexpected EOF")
	}
	return t.current
}

func (t *ParseTree) peek() Token {
	if t.peeked != nil {
		return t.peeked
	}

	p, err := t.tz.ReadToken()
	if err != nil && !errors.Is(err, io.EOF) {
		t.errorf("Unknown Token: %v", err)
	}
	t.peeked = p

	return p
}

func isTokenType(tk Token, tt TokenType) bool {
	return tk != nil && tk.Type() == tt
}

func isTokenOne(tk Token, tt TokenType, val string) bool {
	return isTokenType(tk, tt) && tk.GetValue() == val
}

func isTokenAny(tk Token, tt TokenType, vals ...string) bool {
	if isTokenType(tk, tt) {
		for _, v := range vals {
			if tk.GetValue() == v {
				return true
			}
		}
	}
	return false
}

func (t *ParseTree) feed() Token {
	return t.next()
}

func (t *ParseTree) feedToken(tt TokenType, val string) Token {
	tk := t.next()
	if tk.Type() != tt {
		t.errorf("Unexpexted token. Expected %v %s", tt, tk.GetValue())
	}
	if val != "" && tk.GetValue() != val {
		t.errorf("Unexpected token value. Expected %s; Got: %s", val, tk.GetValue())
	}
	return tk
}

// type:'int'|'char'|'boolean'|className
func (t *ParseTree) feedType() Token {
	tk := t.next()
	switch tk.Type() {
	case TokenKeyword:
		val := tk.GetValue()
		if val != "int" && val != "char" && val != "boolean" {
			t.errorf("Unexpected type")
		}
	case TokenIdentifier:
		break
	default:
		t.errorf("Unexpected Token for Var type")
	}

	return tk
}

// varDec:'var' type varName (','varName)*';'
func (t *ParseTree) varDec() *VarDecNode {
	t.feedToken(TokenKeyword, "var")
	vd := NewVarDecNode(t.feedType(), t.feedToken(TokenIdentifier, ""))
	for !isTokenOne(t.peek(), TokenSymbol, ";") {
		t.feedToken(TokenSymbol, ",")
		vd.AddId(t.feedToken(TokenIdentifier, ""))
	}
	t.feedToken(TokenSymbol, ";")
	return vd
}

// 'let'varName ('['expression']')?'='expression';'
func (t *ParseTree) letStatement() *LetStatementNode {
	t.feedToken(TokenKeyword, "let")
	varNameToken := t.feedToken(TokenIdentifier, "")

	var arrExpr *ExpressionNode
	if p := t.peek(); isTokenOne(p, TokenSymbol, "[") {
		t.feed() //feed [
		arrExpr = t.expression()
		t.feedToken(TokenSymbol, "]")
	}
	t.feedToken(TokenSymbol, "=")
	valExpr := t.expression()
	t.feedToken(TokenSymbol, ";")
	return NewLetStatementNode(varNameToken, arrExpr, valExpr)
}

// term (op term)*
func (t *ParseTree) expression() *ExpressionNode {
	term := t.term()
	en := NewExpressionNode(term)

	p := t.peek()
	for isTokenAny(p, TokenSymbol, "+", "-", "*", "/", "&", "|", "<", ">", "=") {
		opToken := t.feed()
		nextTerm := t.term()
		en.AddOpTerm(opToken, nextTerm)
		p = t.peek()
	}
	return en
}

// term:  integerConstant | stringConstant | keywordConstant | varName | varName'['expression']'|
// subroutineCall |'('expression')'| unaryOp term
func (t *ParseTree) term() *TermNode {
	p := t.peek()
	var tn *TermNode
	switch {
	// integerConstant
	case isTokenType(p, TokenIntegerConst):
		intToken := t.feed()
		tn = NewConstTermNode(intToken)
	// stringConstant
	case isTokenType(p, TokenStringConst):
		strToken := t.feed()
		tn = NewConstTermNode(strToken)
	// keywordConstant
	case isTokenAny(p, TokenKeyword, "true", "false", "null", "this"):
		kwToken := t.feed()
		tn = NewConstTermNode(kwToken)
	// unaryOp term
	case isTokenAny(p, TokenSymbol, "-", "~"):
		unOpTk := t.feed()
		childTerm := t.term()
		tn = NewUnaryTermNode(unOpTk, childTerm)
	case isTokenOne(p, TokenSymbol, "("):
		t.feed()
		expr := t.expression()
		tn = NewExpressionTermNode(expr)
		t.feedToken(TokenSymbol, ")")
	//varName | varName'['expression']'
	case isTokenType(p, TokenIdentifier):
		varName := t.feed()
		pn := t.peek() // peek one more
		if isTokenOne(pn, TokenSymbol, "[") {
			t.feed() // feed [
			expr := t.expression()
			tn = NewArrayTermNode(varName, expr)
			t.feedToken(TokenSymbol, "]")
		} else {
			tn = NewVarTermNode(varName)
		}
	default:
		t.errorf("Token is not a term: %v %s", p.Type(), p.GetValue())
	}

	return tn
}
